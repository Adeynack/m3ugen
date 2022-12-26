use std::fs;

use eyre::Result;
use serde::{Deserialize, Serialize};

#[derive(Debug, Serialize, Deserialize)]
pub struct Configuration {
    #[serde(skip_deserializing)]
    pub config: Option<String>,

    pub debug: bool,

    pub verbose: bool,

    #[serde(default = "String::new")]
    pub output: String,

    #[serde(default = "Vec::new")]
    pub scan: Vec<String>,

    #[serde(default = "Vec::new")]
    pub extensions: Vec<String>,

    pub randomize: bool,

    pub maximum: Option<i64>,

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
