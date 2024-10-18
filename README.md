# BookWeWork

wework is a CLI tool for booking WeWork spaces.

## Installation

You can install wework directly from GitHub using pip:


pip install git+https://github.com/dvcrn/wework-cli.git


## Usage

After installation, you can use the `wework` command in your terminal:


wework [action] [date] [options]


Available actions:
- `book`: Book a WeWork space
- `spaces`: List available spaces
- `locations`: List WeWork locations in a city

Options:
- `--location-uuid`: Location UUID for booking (required for 'book' action)
- `--city`: City name (required for 'locations' action)
- `--username`: Username (required)
- `--password`: Password (required)

Examples:

1. List locations in a city:
   
   wework locations 2023-06-01 --city "New York" --username your_username --password your_password
   

2. List available spaces for a date:
   
   wework spaces 2023-06-01 --location-uuid YOUR_LOCATION_UUID --username your_username --password your_password
   

3. Book a space:
   
   wework book 2023-06-01 --location-uuid YOUR_LOCATION_UUID --username your_username --password your_password
   

Note: You can book multiple dates by using a date range or comma-separated list:
   
   wework book 2023-06-01~2023-06-05 --location-uuid YOUR_LOCATION_UUID --username your_username --password your_password
   wework book 2023-06-01,2023-06-03,2023-06-05 --location-uuid YOUR_LOCATION_UUID --username your_username --password your_password
   

You can also use the `--city` option with the `spaces` action to list available spaces in a specific city:

wework spaces 2023-06-01 --city "New York" --username your_username --password your_password


For more information on available options, use:


wework --help


## Development

To set up the development environment:

1. Clone the repository
2. Install the dependencies: `pip install -r requirements.txt`
3. Make your changes
4. Run tests (if available)
5. Submit a pull request

## License

This project is licensed under the MIT License.
