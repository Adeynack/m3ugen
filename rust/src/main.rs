mod configuration;
mod scan;

use configuration::Configuration;
use eyre::eyre;
use eyre::Result;
use scan::scan;
use std::env;

fn main() -> Result<()> {
    color_eyre::install()?;

    let args: Vec<String> = env::args().collect();
    if args.len() != 2 {
        return Err(eyre!("Expecting the configuration file as argument."));
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
