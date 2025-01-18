# ShutdownManager

`ShutdownManager` is a lightweight and extensible Go module designed to help gracefully manage application shutdown processes. It provides a clean interface for ensuring all resources, such as goroutines, database connections, and external services, are properly closed when your application stops.

---

## Features

- Graceful shutdown handling for long-running applications.
- Ability to register custom shutdown hooks.
- Timeout support to prevent hanging during shutdown.
- Lightweight and easy to integrate with existing projects.

---

## Installation

To add `ShutdownManager` to your project, use:

```bash
go get https://github.com/mohammadJBYaseen/shutdown-manager
```
---

## Contributing

Feel free to fork this repository and submit pull requests. All contributions are welcome!
