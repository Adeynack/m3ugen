use std::fs;

use eyre::Result;
use serde::Deserialize;

#[derive(Debug, Deserialize)]
pub struct Configuration {
    pub verbose: bool,
    #[serde(alias = "output", default = "String::new")]
    pub output_path: String,
    #[serde(alias = "scan", default = "Vec::new")]
    pub scan_folders: Vec<String>,
    #[serde(default = "Vec::new")]
    pub extensions: Vec<String>,
    #[serde(alias = "randomize")]
    pub randomize_list: bool,
    #[serde(alias = "maximum")]
    pub maximum_entries: Option<i64>,
    #[serde(default = "default_as_true")]
    pub detect_duplicates: bool,
}

const fn default_as_true() -> bool {
    true
}

impl Configuration {
    pub fn load_from_file(path: &str) -> Result<Self> {
        let config_content = fs::read_to_string(path)?;
        Ok(serde_yaml::from_str(&config_content)?)
    }
}
