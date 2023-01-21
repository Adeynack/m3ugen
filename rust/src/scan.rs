use crate::configuration::Configuration;
use eyre::eyre;
use eyre::Result;
use std::{collections::HashSet, fs, path::Path};

#[derive(Debug)]
#[allow(clippy::module_name_repetitions)]
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

    pub fn report(&self, config: &Configuration) {
        if config.verbose {
            config.verbose_print("Excluded extensions:");
            self.excluded_extensions.iter().for_each(|ext| config.verbose_print(&format!("  - {ext}")));
        }
    }

    pub fn write_result(&self, config: &Configuration) {
        match config.output {
            Some(_) => todo!(),
            None => self.found_file_paths.iter().for_each(|p| println!("{p}")),
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
        self.configuration.verbose_print("---=== m3u Playlist Generator ===---");

        for folder in &self.configuration.scan {
            self.scan_folder(Path::new(folder)).map_err(|e| eyre!("Unable to scan folder: {}", e))?;
        }

        Ok(())
    }

    fn scan_folder(&mut self, folder_path: &Path) -> Result<()> {
        self.configuration.verbose_print(&format!("Scanning folder {}", folder_path.to_str().unwrap_or("?")));
        let read_dir = fs::read_dir(folder_path).map_err(|e| eyre!("Unable to read directory {:?}: {}", folder_path, e))?;
        for entry in read_dir {
            let path = entry?.path();
            if path.is_dir() {
                self.scan_folder(&path)?;
            } else {
                self.consider_file_path(&path);
            }
        }

        Ok(())
    }

    fn consider_file_path(&mut self, path: &Path) {
        let path_str = path.to_string_lossy().to_string();
        self.configuration.debug_print(&format!("Considering file {path_str}"));
        match path.extension().map(|e| e.to_string_lossy().to_string()) {
            None => self.consider_file_without_extension(path_str),
            Some(extension) => self.consider_file_with_extension(extension, path_str),
        }
    }

    fn consider_file_without_extension(&mut self, path_str: String) {
        if self.configuration.extensions.is_empty() {
            self.result.found_file_paths.push(path_str);
        } else {
            self.configuration.debug_print("Ignoring file without extension");
        }
    }

    fn consider_file_with_extension(&mut self, extension: String, path_str: String) {
        if self.configuration.extensions.contains(&extension) {
            self.result.found_file_paths.push(path_str);
        } else {
            self.configuration.debug_print(&format!("Ignoring file with extension '{extension}'"));
            self.result.excluded_extensions.insert(extension);
        }
    }
}
