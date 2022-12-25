use crate::configuration::Configuration;
use eyre::eyre;
use eyre::Result;
use std::{collections::HashSet, fs, path::Path};

#[derive(Debug)]
pub struct ScanResult {
    pub found_file_paths: Vec<String>,
    pub excluded_extensions: HashSet<String>,
}

impl ScanResult {
    pub fn new() -> Self {
        Self {
            found_file_paths: Vec::new(),
            excluded_extensions: HashSet::new(),
        }
    }
}

pub fn scan(configuration: &Configuration) -> Result<ScanResult> {
    let mut scan_session = Scan {
        result: ScanResult::new(),
        configuration,
    };
    scan_session.start()?;
    Ok(scan_session.result)
}

struct Scan<'a> {
    configuration: &'a Configuration,
    result: ScanResult,
}

impl Scan<'_> {
    fn start(&mut self) -> Result<()> {
        for folder in &self.configuration.scan_folders {
            self.scan_folder(Path::new(folder))
                .map_err(|e| eyre!("Unable to scan folder: {}", e))?;
        }

        Ok(())
    }

    fn scan_folder(&mut self, folder_path: &Path) -> Result<()> {
        let read_dir = fs::read_dir(folder_path)
            .map_err(|e| eyre!("Unable to read directory {:?}: {}", folder_path, e))?;
        for entry in read_dir {
            let path = entry?.path();
            if path.is_dir() {
                self.scan_folder(&path)?;
            } else {
                self.consider_file_path(&path)?;
            }
        }

        Ok(())
    }

    fn consider_file_path(&mut self, file_path: &Path) -> Result<()> {
        let file_path_str = file_path.to_str().unwrap();
        self.result.found_file_paths.push(file_path_str.to_string());
        Ok(())
    }
}
