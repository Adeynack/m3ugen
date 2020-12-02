mod configuration;

use configuration::Configuration;
use simple_error::SimpleError;
use std::fs;

fn main() {
    println!("Hello, world!");
    // let config_file_content = fs::read_to_string("./doc/config_example.yaml").expect("Error reading configuration file.");
    // println!("config_file_content: {}", config_file_content);
    // let config: Configuration = serde_yaml::from_str(&config_file_content).expect("Error deserializing configuration.");
    let config =
        load_configuration("./doc/config_example.yaml").expect("Error loading configuration");
    println!("Configuration: {:?}", config);
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
