#[macro_use]
extern crate simple_error;

mod configuration;
mod scan;

use configuration::Configuration;
use scan::scan;
use std::{env, error::Error, process};

fn main() {
    if let Err(err) = main_or_error() {
        eprintln!("{}", err.to_string());
        process::exit(1);
    }
}

fn main_or_error() -> Result<(), Box<dyn Error>> {
    println!("---=== m3u Playlist Generator ===---");

    let args: Vec<String> = env::args().collect();
    if args.len() != 2 {
        bail!("Expecting the configuration file as argument.")
    }

    let config_file_path = &args[1];
    let config = Configuration::load_from_file(config_file_path)?;

    let scan_result = scan(&config)?;
    println!("Files found:");
    scan_result
        .found_file_paths
        .iter()
        .for_each(|f| println!("  - {}", f));

    Ok(())
}
