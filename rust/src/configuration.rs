use std::fs;

use clap::Parser;
use derive_builder::Builder;
use eyre::eyre;
use eyre::Result;
use serde::{Deserialize, Serialize};

#[derive(Debug, Serialize, Deserialize, Parser, Default, Builder)]
#[command()]
#[builder(default)]
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
    pub maximum: Option<u64>,

    /// Perform duplicate detection. By default: true.
    #[arg(short, long)]
    pub detect_duplicates: bool,
}

impl Configuration {
    /// Loads the configuration from the CLI, and if needed from a YAML file.
    pub fn load() -> Result<Configuration> {
        let cli = Self::parse().normalize();
        let config = match cli.config {
            Some(ref path) => Self::load_from_file(path)?.normalize().merge(cli),
            None => cli,
        };
        if config.debug {
            config.debug_print("Loaded configuration:");
            config.debug_print(&serde_json::to_string_pretty(&config)?);
        }
        Ok(config)
    }

    pub fn normalize(mut self) -> Self {
        self.extensions = self.extensions.iter().map(|e| e.to_lowercase()).collect();
        self
    }

    /// Loads the configuration from a YAML file.
    pub fn load_from_file(path: &str) -> Result<Self> {
        let config_content = fs::read_to_string(path).map_err(|e| eyre!("Unable to load configuration file at {path} ({e})"))?;
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
