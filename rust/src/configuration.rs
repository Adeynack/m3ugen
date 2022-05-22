use std::{error::Error, fs};

use serde::Deserialize;

#[derive(Debug, Deserialize)]
pub struct Configuration {
    #[serde(default = "default_as_false")]
    pub verbose: bool,
    #[serde(alias = "output", default = "String::new")]
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

impl Configuration {
    pub fn load_from_file(path: &str) -> Result<Configuration, Box<dyn Error>> {
        let config_content = fs::read_to_string(path)?;
        Ok(serde_yaml::from_str(&config_content)?)
    }
}
