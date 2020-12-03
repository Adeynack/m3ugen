mod configuration;

use configuration::Configuration;
use simple_error::SimpleError;
use std::{env, fs, process};

fn main() {
    if let Err(err) = main_or_error() {
        eprintln!("{}", err.to_string());
        process::exit(1);
    }
}

fn main_or_error() -> Result<(), SimpleError> {
    println!("---=== m3u Playlist Generator ===---");

    let args: Vec<String> = env::args().collect();
    if args.len() != 2 {
        return Err(SimpleError::new(
            "Expecting the configuration file as argument.",
        ));
    }
    let config_filename = &args[1];
    println!("Using configuration file at {}", config_filename);

    let config = load_configuration(config_filename)?;
    println!("Configuration: {:?}", config);

    Ok(())
}

fn load_configuration(configuration_file_path: &str) -> Result<Configuration, SimpleError> {
    let config_content = fs::read_to_string(configuration_file_path).map_err(|err| {
        SimpleError::new(format!(
            "Unable to read content of the configuration file: {}",
            err
        ))
    })?;
    serde_yaml::from_str::<Configuration>(&config_content).map_err(|err| {
        SimpleError::new(format!("Unable to parse configuration from YAML: {}", err))
    })
}
