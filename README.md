# Logger Package

A structured, high-performance logger with colored console output and optional file logging, built on `zap.Logger`.

## Features
- **Colored Logging**: Log levels and messages are color-coded for easy readability.
- **Structured Logging**: Add fields and structured data to logs.
- **File Logging**: Optionally log to a file in JSON format.
- **Convenience Methods**: Helper methods for common logging scenarios (e.g., `Success`, `Progress`, `Warning`, `Failure`).
- **Custom Encoders**: Separate encoders for console (colored) and file (JSON) outputs.
- **Thread-Safe**: Built on `zap.Logger` for high performance and thread safety.
