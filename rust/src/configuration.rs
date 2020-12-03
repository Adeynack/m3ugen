use std::fs;

use serde::Deserialize;
use simple_error::SimpleError;

#[derive(Debug, Deserialize)]
pub struct Configuration {
    #[serde(default = "default_as_false")]
    pub verbose: bool,
    #[serde(alias = "output", default = "default_output_path")]
    pub output_path: String,
    #[serde(alias = "scan", default = "Vec::new")]
    pub scan_folders: Vec<String>,
    #[serde(default = "Vec::new")]
    pub extensions: Vec<String>,
    #[serde(default = "default_as_false")]
    pub randomize_list: bool,
    pub maximum_entries: Option<i64>,
    #[serde(default = "default_as_true")]
    pub detect_duplicates: bool,
}

fn default_as_true() -> bool {
    true
}
fn default_as_false() -> bool {
    false
}
fn default_output_path() -> String {
    String::new()
}

impl Configuration {
    pub fn load_from_file(path: &str) -> Result<Configuration, SimpleError> {
        let config_content = fs::read_to_string(path).map_err(|err| {
            SimpleError::new(format!(
                "Unable to read content of the configuration file: {}",
                err
            ))
        })?;
        serde_yaml::from_str::<Configuration>(&config_content).map_err(|err| {
            SimpleError::new(format!("Unable to parse configuration from YAML: {}", err))
        })
    }
}
