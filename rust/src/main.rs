mod configuration;
mod scan;

use configuration::Configuration;
use eyre::Result;
use scan::scan;

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

    let config = Configuration::load()?;
    let scan_result = scan(&config)?;

    eprintln!("Files found:");
    scan_result
        .found_file_paths
        .iter()
        .for_each(|f| eprintln!("  - {f}"));

    Ok(())
}
