mod configuration;
mod scan;

use std::rc::Rc;

use configuration::Configuration;
use eyre::Result;
use scan::Scan;

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

    let config = Rc::new(Configuration::load()?);
    let scan_result = Scan::scan(Rc::clone(&config))?;
    scan_result.report(&config);
    scan_result.write_result(&config);
    Ok(())
}
