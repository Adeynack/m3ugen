use std::{fs, path::Path};

use simple_error::SimpleError;

use crate::configuration::Configuration;

#[derive(Debug)]
pub struct ScanResult {
    pub found_file_paths: Vec<String>,
    pub excluded_extensions: Vec<String>,
}

impl ScanResult {
    pub fn new() -> Self {
        Self {
            found_file_paths: vec![],
            excluded_extensions: vec![],
        }
    }
}

pub fn scan(configuration: &Configuration) -> Result<ScanResult, SimpleError> {
    let mut result = ScanResult::new();

    for folder in &configuration.scan_folders {
        scan_folder(&configuration, Path::new(folder), &mut result)
            .map_err(|err| SimpleError::new(format!("Unable to scan folder: {}", err)))?;
    }

    Ok(result)
}

fn scan_folder(
    configuration: &Configuration,
    folder_path: &Path,
    result: &mut ScanResult,
) -> Result<(), SimpleError> {
    let read_dir = fs::read_dir(folder_path).map_err(|err| {
        SimpleError::new(format!(
            "Unable to read directory {:?}: {}",
            folder_path, err
        ))
    })?;
    for entry in read_dir {
        let entry = entry.map_err(|err| SimpleError::new(err.to_string()))?;
        let path = entry.path();
        if path.is_dir() {
            scan_folder(&configuration, &path, result)?;
        } else {
            consider_file_path(&configuration, &path, result)?;
        }
    }

    Ok(())
}

fn consider_file_path(
    _configuration: &Configuration,
    file_path: &Path,
    result: &mut ScanResult,
) -> Result<(), SimpleError> {
    let file_path_str = file_path.to_str().ok_or_else(|| {
        SimpleError::new(format!("Path didn't extract to string: {:?}", file_path))
    })?;
    result.found_file_paths.push(file_path_str.to_string());
    Ok(())
}
