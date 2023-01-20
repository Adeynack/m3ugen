use std::fs;

use clap::Parser;
use eyre::Result;
use serde::{Deserialize, Serialize};

#[derive(Debug, Serialize, Deserialize, Parser)]
#[command()]
pub struct Configuration {
    /// Path to a YAML configuration file.
    #[serde(skip_deserializing)]
    #[arg(short, long)]
    pub config: Option<String>,

    /// Display debug information.
    #[arg(long)]
    pub debug: bool,

    /// Display progress and information to the console.
    #[arg(short, long)]
    pub verbose: bool,

    /// File path of the output playlist. If none, will output to the console.
    #[arg(short, long)]
    pub output: Option<String>,

    /// Folders to scan. By default, scans the current folder.
    #[serde(default = "Vec::new")]
    pub scan: Vec<String>,

    /// File extensions to include in playlist. By default, takes all.
    #[serde(default = "Vec::new")]
    #[arg(short, long)]
    pub extensions: Vec<String>,

    /// Randomize the generated playlist.
    #[arg(short, long)]
    pub randomize: bool,

    /// Limits the maximum files in the playlist.
    #[arg(short, long)]
    pub maximum: Option<i64>,

    /// Perform duplicate detection. By default: true.
    #[serde(default = "default_as_true")]
    #[arg(short, long)]
    pub detect_duplicates: bool,
}

const fn default_as_true() -> bool {
    true
}

impl Configuration {
    /// Loads the configuration from a YAML file.
    pub fn load_from_file(path: &str) -> Result<Self> {
        let config_content = fs::read_to_string(path)?;
        Ok(serde_yaml::from_str(&config_content)?)
    }

    /// Merges another configuration into the current.
    pub fn merge(mut self, other: Configuration) -> Self {
        self.config = other.config.or(self.config);
        self.debug |= other.debug;
        self.detect_duplicates |= other.detect_duplicates;
        self.extensions.extend(other.extensions);
        self.maximum = other.maximum.or(self.maximum);
        self.output = other.output.or(self.output);
        self.randomize |= other.randomize;
        self.scan.extend(other.scan);
        self.verbose |= other.verbose;
        self
    }

    pub fn verbose_print(&self, message: &str) {
        if self.verbose {
            eprintln!("{message}");
        }
    }

    pub fn debug_print(&self, message: &str) {
        if self.debug {
            eprintln!("{message}");
        }
    }
}
