# BookWeWork

wework is a CLI tool for booking WeWork spaces.

## Installation

You can install wework directly from GitHub using pip:

```
pip install git+https://github.com/dvcrn/wework-cli.git
```

## Usage

After installation, you can use the `wework` command in your terminal:

```
wework [action] [date] [options]
```

Available actions:
- `book`: Book a WeWork space
- `spaces`: List available spaces
- `locations`: List WeWork locations in a city

Options:
- `--location-uuid`: Location UUID for booking (required for 'book' action)
- `--city`: City name (required for 'locations' action)
- `--space-name`: Name of a specific space (optional)

Examples:

1. List locations in a city:
   ```
   wework locations 2023-06-01 --city "New York"
   ```

2. List available spaces for a date:
   ```
   wework spaces 2023-06-01 --location-uuid YOUR_LOCATION_UUID
   ```

3. Book a space:
   ```
   wework book 2023-06-01 --location-uuid YOUR_LOCATION_UUID
   ```

For more information on available options, use:

```
wework --help
```

## Development

To set up the development environment:

1. Clone the repository
2. Install the dependencies: `pip install -r requirements.txt`
3. Make your changes
4. Run tests (if available)
5. Submit a pull request

## License

This project is licensed under the MIT License.
