use std::{collections::HashSet, error::Error, fs, path::Path};

use simple_error::SimpleError;

use crate::configuration::Configuration;

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

pub fn scan(configuration: &Configuration) -> Result<ScanResult, SimpleError> {
    let mut scan_session = Scan {
        result: ScanResult::new(),
    };
    scan_session.start(configuration)?;
    Ok(scan_session.result)
}

struct Scan {
    result: ScanResult,
}

impl Scan {
    fn start(&mut self, configuration: &Configuration) -> Result<(), SimpleError> {
        for folder in &configuration.scan_folders {
            self.scan_folder(&configuration, Path::new(folder))
                .map_err(|e| simple_error!("Unable to scan folder: {}", e))?;
        }

        Ok(())
    }

    fn scan_folder(
        &mut self,
        configuration: &Configuration,
        folder_path: &Path,
    ) -> Result<(), Box<dyn Error>> {
        let read_dir = fs::read_dir(folder_path)
            .map_err(|e| simple_error!("Unable to read directory {:?}: {}", folder_path, e))?;
        for entry in read_dir {
            let path = entry?.path();
            if path.is_dir() {
                self.scan_folder(&configuration, &path)?;
            } else {
                self.consider_file_path(&configuration, &path)?;
            }
        }

        Ok(())
    }

    fn consider_file_path(
        &mut self,
        _configuration: &Configuration,
        file_path: &Path,
    ) -> Result<(), SimpleError> {
        let file_path_str = file_path.to_str().unwrap();
        self.result.found_file_paths.push(file_path_str.to_string());
        Ok(())
    }
}
