mod configuration;
mod scan;

use configuration::Configuration;
use eyre::eyre;
use eyre::Result;
use scan::scan;
use std::env;

#[warn(clippy::suspicious)]
#[warn(clippy::complexity)]
#[warn(clippy::perf)]
#[warn(clippy::style)]
#[warn(clippy::pedantic)]
#[warn(clippy::cargo)]
#[warn(clippy::nursery)]
#[warn(clippy::unwrap_used)]
#[warn(clippy::expect_used)]
#[allow(clippy::implicit_return)]
fn main() -> Result<()> {
    color_eyre::install()?;

    let args: Vec<String> = env::args().collect();
    let config = args.get(1).map_or(
        Err(eyre!("Expecting the configuration file as argument.")),
        |config_file_path| Configuration::load_from_file(config_file_path),
    )?;
    if config.debug {
        if let Ok(pretty_config) = serde_json::to_string_pretty(&config) {
            eprintln!("Loaded configuration:");
            eprintln!("{pretty_config}");
        }
    }

    let scan_result = scan(&config)?;

    eprintln!("Files found:");
    scan_result
        .found_file_paths
        .iter()
        .for_each(|f| eprintln!("  - {f}"));

    Ok(())
}
