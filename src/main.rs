use std::error::Error;
use std::fs::File;
use std::io::BufReader;
use std::io::{stdin, stdout, Write};
use std::thread::JoinHandle;
use std::time::Duration;
use std::{fmt, thread};

use crossterm::event::KeyModifiers;
use futures::channel::oneshot::{self, Cancellation};
use rodio::source::{SineWave, Source};
use rodio::{Decoder, OutputStream, Sink};

use midir::{Ignore, MidiInput, MidiOutput};

use std::io;

use tokio::time;
use tokio_util::task::TaskTracker;

use color_eyre::Result;
use futures::{FutureExt, StreamExt};
use ratatui::crossterm::event::{self, Event, EventStream, KeyCode, KeyEvent, KeyEventKind};
use ratatui::prelude::*;
use ratatui::symbols::*;
use ratatui::widgets::*;
use ratatui::DefaultTerminal;
/*
use ratatui::{
    buffer::Buffer,
    crossterm::event::{self, Event, KeyCode, KeyEvent, KeyEventKind},
    layout::{Constraint, Layout, Margin, Position},
    layout::{Direction, Rect},
    style::{Color, Modifier, Style, Stylize},
    symbols::border,
    text::{Line, Span, Text},
    widgets::{Bar, BarChart, BarGroup, Block, Gauge, List, ListItem, Paragraph, Widget},
    DefaultTerminal, Frame,
};
*/

use better_panic::Settings;

use log::error;

use warp::Filter;

use tokio::sync::broadcast;

// mod tui;

#[tokio::main]
async fn main() -> Result<()> {
    color_eyre::install()?;
    println!("Hello, world!");
    let (tx, mut rx1) = broadcast::channel(16);
    list_ports();

    // _stream must live as long as the sink
    let (_stream, stream_handle) = OutputStream::try_default().unwrap();
    let sink = Sink::try_new(&stream_handle).unwrap();

    // Add a dummy source of the sake of the example.
    let source = SineWave::new(440.0)
        .take_duration(Duration::from_secs_f32(0.25))
        .amplify(0.20);
    sink.append(source);

    let tracker = TaskTracker::new();

    let hello = warp::path!("hello" / String).map(|name| format!("Hello, {}!", name));

    let (shutdown_tx, shutdown_rx) = tokio::sync::oneshot::channel::<()>();
    // let server = warp::serve(hello).try_bind(([127, 0, 0, 1], 3030));
    let tx_clone = tx.clone();
    let server_result =
        warp::serve(hello).try_bind_with_graceful_shutdown(([127, 0, 0, 1], 3030), async move {
            // shutdown_rx.await.ok();
            let mut rx = tx_clone.subscribe();
            let _ = rx.recv().await;
            println!("Web server stopping...");
        });
    // let server_handle = tracker.spawn(server);
    /*
    match server_result {
        Ok((addr, server)) => {
            let driver = tracker.spawn(server);
            Ok((driver, addr))
        }
        Err(e) => Err(format!("Failed to bind socket for webserver: {:?}", e)),
    }
    */

    match server_result {
        Ok((addr, server)) => {
            let _ = tracker.spawn(server);
            ()
        }
        Err(e) => (),
    }

    let terminal = ratatui::init();
    // let app_result = App::default().run(terminal).await;
    let app = App::new();
    let tui_tx_clone = tx.clone();
    let tui_handle = tracker.spawn(async move {
        // let app_result = App::new().run(shutdown_tx, terminal).await;
        let app_result = App::new().run(terminal).await;
        match app_result {
            Ok(()) => {
                print!("\nOK\n");
                // server_handle.abort();
                // shutdown_tx.send(());
                tui_tx_clone.send(()).unwrap();
            }
            Err(err) => {
                print!("\nError {}\n", err);
                ()
            }
        };
    });
    let tx_clone = tx.clone();
    /*
    tracker.spawn(async move {
        match tui_handle.await {
            Ok(()) => {
                _ = tx_clone.send(());
            }
            Err(e) => {
                tokio::time::sleep(Duration::from_secs(5)).await;
                ratatui::restore();
                log::warn!("No UI task.: {}", e);
            }
        }
    });
    */
    tracker.spawn(async move {
        let mut rx = tx_clone.subscribe();
        tokio::select! {
            _ = tokio::signal::ctrl_c() => (),
            _ = rx.recv() => ()
        }
        ratatui::restore();
        print!("TUI off: {}\n", tui_handle.is_finished());
        println!("Quitting...");
        tx.send(());
    });
    tracker.close();
    tracker.wait().await;
    // let may_panic = async { App::default().run(terminal) };
    // let async_result = may_panic.catch_unwind().await;
    /*
    loop {
        terminal.draw(draw).expect("failed to draw frame");
        if matches!(event::read().expect("failed to read event"), Event::Key(_)) {
            break;
        }
    }
    */
    ratatui::restore();

    // The sound plays in a separate thread. This call will block the current thread until the sink
    // has finished playing all its queued sounds.
    // sink.sleep_until_end();

    println!("Done.");
    // app_result
    Ok(())
}

pub fn initialize_panic_handler() -> Result<()> {
    let (panic_hook, eyre_hook) = color_eyre::config::HookBuilder::default()
        .panic_section(format!(
            "This is a bug. Consider reporting it at {}",
            env!("CARGO_PKG_REPOSITORY")
        ))
        .display_location_section(true)
        .display_env_section(true)
        .into_hooks();
    eyre_hook.install()?;
    std::panic::set_hook(Box::new(move |panic_info| {
        /*
        if let Ok(t) = crate::tui::Tui::new() {
            if let Err(r) = t.exit() {
                error!("Unable to exit Terminal: {:?}", r);
            }
        }
        */

        let msg = format!("{}", panic_hook.panic_report(panic_info));
        #[cfg(not(debug_assertions))]
        {
            eprintln!("{msg}");
            use human_panic::{handle_dump, print_msg, Metadata};
            let author = format!("authored by {}", env!("CARGO_PKG_AUTHORS"));
            let support = format!(
                "You can open a support request at {}",
                env!("CARGO_PKG_REPOSITORY")
            );
            let meta = Metadata::new(env!("CARGO_PKG_NAME"), env!("CARGO_PKG_VERSION"))
                .authors(author)
                .support(support);

            let file_path = handle_dump(&meta, panic_info);
            print_msg(file_path, &meta)
                .expect("human-panic: printing error message to console failed");
        }
        // log::error!("Error: {}", strip_ansi_escapes::strip_str(msg));

        #[cfg(debug_assertions)]
        {
            // Better Panic stacktrace that is only enabled when debugging.
            better_panic::Settings::auto()
                .most_recent_first(false)
                .lineno_suffix(true)
                .verbosity(better_panic::Verbosity::Full)
                .create_panic_handler()(panic_info);
        }

        // std::process::exit(libc::EXIT_FAILURE);
    }));
    Ok(())
}
// #[derive(Debug, Default)]
pub struct App {
    counter: u8,
    exit_counter: u8,
    FPS: f64,
    exit: bool,
}

impl App {
    const FPS: f64 = 60.0;
    pub fn new() -> Self {
        Self {
            counter: 0,
            exit_counter: 0,
            FPS: 60.0,
            exit: false,
        }
    }
    /// runs the application's main loop until the user quits
    pub async fn run(
        mut self,
        // shutdown: tokio::sync::oneshot::Sender<()>,
        mut terminal: DefaultTerminal,
    ) -> Result<()> {
        self.FPS = Self::FPS;
        let period = Duration::from_secs_f64(1.0 / Self::FPS);
        let mut interval = tokio::time::interval(period);
        let mut events = EventStream::new();
        while !self.exit {
            tokio::select! {
                _ = interval.tick() => {
                    if self.exit_counter > 0 {
                        self.exit_counter -= 1;
                    }
                    terminal.draw(|frame| self.draw(frame))?;
                },
                Some(Ok(event)) = events.next() => self.handle_event(&event)?,
            }
        }
        // shutdown.send(());
        Ok(())
    }

    fn draw(&self, frame: &mut Frame) {
        frame.render_widget(self, frame.area());
    }

    fn handle_event(&mut self, event: &Event) -> Result<()> {
        if let Event::Key(key) = event {
            if key.kind == KeyEventKind::Press {
                self.handle_key_event(key)?;
            }
        }
        Ok(())
    }

    fn handle_key_event(&mut self, key_event: &KeyEvent) -> Result<()> {
        let has_ctrl = key_event.modifiers.contains(KeyModifiers::CONTROL);
        match key_event.code {
            KeyCode::Char('q') if has_ctrl => self.exit(),
            KeyCode::Char('d') if has_ctrl => self.exit(),
            KeyCode::Char('c') if has_ctrl => self.exit(),
            KeyCode::Left => self.decrement_counter()?,
            KeyCode::Right => self.increment_counter()?,
            _ => (),
        }
        Ok(())
    }
    fn exit(&mut self) {
        if self.exit_counter > 0 {
            self.exit = true;
        } else {
            self.exit_counter = 100;
        }
    }

    fn decrement_counter(&mut self) -> Result<()> {
        self.counter -= 1;
        Ok(())
    }

    fn increment_counter(&mut self) -> Result<()> {
        self.counter += 1;
        Ok(())
    }
}

impl Widget for &App {
    fn render(self, area: Rect, buf: &mut Buffer) {
        let vertical = &Layout::vertical([Constraint::Min(5), Constraint::Length(3)]);
        let rects = vertical.split(area);
        let title = Line::from(" Counter App Tutorial ".bold());
        let instructions = Line::from(vec![
            " Decrement ".into(),
            "<Left>".blue().bold(),
            " Increment ".into(),
            "<Right>".blue().bold(),
            " Quit ".into(),
            "<Ctrl+Q> ".blue().bold(),
        ]);
        let quit_gauge = Gauge::default()
            .percent(self.exit_counter.into())
            .label(format!("{:.1}s", f64::from(self.exit_counter) / self.FPS));
        let quit_info = Line::from(vec![
            " Press again to quit. ".red().bold(),
            "[            ] ".into(),
        ]);
        let quit_size = quit_info.width();

        let mut block = Block::bordered()
            .title(title.centered())
            .title_bottom(instructions.right_aligned())
            .border_set(border::THICK);
        if self.exit_counter > 0 {
            block = block.title_bottom(quit_info.left_aligned());
        }

        let counter_text = Text::from(vec![Line::from(vec![
            "Value: ".into(),
            self.counter.to_string().yellow(),
        ])]);

        Paragraph::new(counter_text)
            .centered()
            .block(block)
            .render(area, buf);

        let quit_length = quit_size as u16;
        log::info!("QL: {}\n", quit_length);
        let quit_area = Rect::new(area.left() + quit_length - 12, area.bottom() - 1, 10, 1);
        if self.exit_counter > 0 {
            quit_gauge.render(area.intersection(quit_area), buf);
        }
    }
}

fn list_ports() -> Result<(), Box<dyn Error>> {
    let mut midi_in = MidiInput::new("midir test input")?;
    midi_in.ignore(Ignore::None);
    let midi_out = MidiOutput::new("midir test output")?;

    println!("Available input ports:");
    for (i, p) in midi_in.ports().iter().enumerate() {
        println!("{}: {}", i, midi_in.port_name(p)?);
    }

    println!("\nAvailable output ports:");
    for (i, p) in midi_out.ports().iter().enumerate() {
        println!("{}: {}", i, midi_out.port_name(p)?);
    }

    Ok(())
}
