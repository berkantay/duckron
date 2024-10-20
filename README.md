# Duckron

Duckron is a Go-based tool for managing DuckDB snapshots. It periodically creates snapshots of a DuckDB database and stores them in a specified directory.

## Table of Contents

- [Installation](#installation)
- [Configuration](#configuration)
- [Usage](#usage)
- [Project Structure](#project-structure)
- [License](#license)

## Installation

1. Clone the repository:

   ```sh
   git clone https://github.com/yourusername/duckron.git
   cd duckron
   ```

2. Install dependencies:
   ```sh
   go mod tidy
   ```

## Configuration

Create a `duckron.yaml` file in the root directory with the following structure:

```yaml
path: "my_database.duckdb"
interval: 3600
destinationPath: "backups"
snapshotFormat: "parquet"
```
