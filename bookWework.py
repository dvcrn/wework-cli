import requests
from datetime import datetime, timedelta
import argparse
import sys
from weworkauth import WeWorkAuth


class SharedWorkspaceResponse:
    def __init__(self, data):
        self.limit = data['limit']
        self.offset = data['offset']
        self.workspaces = [Workspace(w) for w in data['getSharedWorkspaces']['workspaces']]

class LocationsByGeoResponse:
    def __init__(self, data):
        self.locations = [GeoLocation(l) for l in data["locationsByGeo"]]


class BookSpaceResponse:
    def __init__(self, data):
        self.booking_processing_status = data['bookingProcessingStatus']
        self.errors = data['errors']
        self.is_errored = data['isErrorred']
        self.reservation_uuid = data['reservationUUID']


class GeoLocation:
    def __init__(self, data):
        self.uuid = data['uuid']
        self.name = data['name']
        self.latitude = data['latitude']
        self.longitude = data['longitude']
        self.address = Address(data['address'])
        self.time_zone = data['timeZone']
        self.distance = data['distance']
        self.brand_name = data['brandName']
        self.has_third_party_display = data['hasThirdPartyDisplay']
        self.image = data['image']
        self.is_migrated = data['isMigrated']

    

class Workspace:
    def __init__(self, data):
        self.uuid = data['uuid']
        self.inventory_uuid = data['inventoryUuid']
        self.image_url = data['imageURL']
        self.header_image_url = data['headerImageUrl']
        self.capacity = data['capacity']
        self.credits = data['credits']
        self.location = Location(data['location'])
        self.open_time = data['openTime']
        self.close_time = data['closeTime']
        self.cancellation_policy = data['cancellationPolicy']
        self.operating_hours = [OperatingHours(oh) for oh in data['operatingHours']]
        self.product_price = ProductPrice(data['productPrice'])
        self.seat = Seat(data['seat'])
        self.seats_available = data['seatsAvailable']
        self.order = data['order']
        self.is_vast_coworking = data['isVASTCoworking']
        self.is_affiliate_coworking = data['isAffiliateCoworking']
        self.is_franchise_coworking = data['isFranchiseCoworking']
        self.is_hybrid_space = data['isHybridSpace']

class Location:
    def __init__(self, data):
        self.description = data['description']
        self.support_email = data['supportEmail']
        self.phone_normalized = data['phoneNormalized']
        self.currency = data['currency']
        self.primary_team_member = TeamMember(data['primaryTeamMember'])
        self.amenities = [Amenity(a) for a in data['amenities']]
        self.details = Details(data['details'])
        self.transit_info = TransitInfo(data['transitInfo'])
        self.member_entrance_instructions = data['memberEntranceInstructions']
        self.parking_instructions = data['parkingInstructions']
        self.timezone_offset = data['timezoneOffset']
        self.time_zone_identifier = data['timeZoneIdentifier']
        self.time_zone_win_id = data['timeZoneWinId']
        self.uuid = data['uuid']
        self.name = data['name']
        self.latitude = data['latitude']
        self.longitude = data['longitude']
        self.address = Address(data['address'])
        self.time_zone = data['timeZone']
        self.distance = data['distance']
        self.has_third_party_display = data['hasThirdPartyDisplay']
        self.is_migrated = data['isMigrated']

class TeamMember:
    def __init__(self, data):
        self.name = data['name']
        self.business_title = data['businessTitle']
        self.image_url = data['imageUrl']

class Amenity:
    def __init__(self, data):
        self.uuid = data['uuid']
        self.name = data['name']
        self.highlight = data['highlight']

class Details:
    def __init__(self, data):
        self.has_extended_hours = data['hasExtendedHours']

class TransitInfo:
    def __init__(self, data):
        self.bike = data['bike']
        self.bus = data['bus']
        self.ferry = data['ferry']
        self.freeway = data['freeway']
        self.metro = data['metro']
        self.parking = data['parking']

class Address:
    def __init__(self, data):
        self.line1 = data['line1']
        self.line2 = data['line2']
        self.city = data['city']
        self.state = data['state']
        self.country = data['country']
        self.zip = data['zip']

class OperatingHours:
    def __init__(self, data):
        self.day_of_week = data['dayOfWeek']
        self.day = data['day']
        self.open = data['open']
        self.close = data['close']
        self.is_closed = data['isClosed']

class ProductPrice:
    def __init__(self, data):
        self.uuid = data['uuid']
        self.product_uuid = data['productUuid']
        self.price = Price(data['price'])
        self.rate_unit = data['rateUnit']
        self.half_hour_credit_prices = [HalfHourCreditPrice(hcp) for hcp in data['halfHourCreditPrices']]

class Price:
    def __init__(self, data):
        self.currency = data['currency']
        self.amount = data['amount']

class HalfHourCreditPrice:
    def __init__(self, data):
        self.offset = data['offset']
        self.amount = data['amount']

class Seat:
    def __init__(self, data):
        self.total = data['total']
        self.available = data['available']


class WeWork: 

    def __init__(self, authorization, wework_auth ):
        self.headers = {
            'Referer': 'https://members.wework.com/workplaceone/content2/bookings/desks',
            'Authorization': f'Bearer {authorization}',
            'Accept': 'application/json, text/plain, */*',
            'Content-Type': 'application/json',
            'WeWorkAuth': f'Bearer {wework_auth}',
            'IsKube': 'true',
            'Request-Source': 'MemberWeb/WorkplaceOne/Prod',
            'WeWorkMemberType': '2',
        }

    def do_request(self, method, url, data=None):
        if method.lower() == 'get':
            response = requests.get(url, headers=self.headers)
        elif method.lower() == 'post':
            response = requests.post(url, headers=self.headers, json=data)
        else:
            raise ValueError("Unsupported HTTP method")

        if response.status_code == 200:
            js = response.json()
            if 'isErrorred' in js and js['isErrorred']:
                error_message = js.get('errors', ['Unknown error'])[0]
                error_status = js.get('errorStatusCode', 'unknown_error')
                raise Exception(f"Request failed: {error_message} (Status: {error_status})")

            return js
        else:
            print(f"Request failed with status code: {response.status_code}")
            return None

    
    def mxg_poll_quote(self, date_str, location_id, reservable_id):
        url = 'https://members.wework.com/workplaceone/api/ext-booking/mxg-poll-quote?APIen=0'
        data = {
            "reservableId": reservable_id,
            "type": 4,
            "creditsUsed": 0,
            "currency": "com.wework.credits",
            "TriggerCalenderEvent": True,
            "mailData": {
                "dayFormatted": date_str,
                "startTimeFormatted": "08:30:00 AM",
                "endTimeFormatted": "20:00:00 PM",
                "floorAddress": "",
                "locationAddress": "6-12-18 Jingumae",
                "creditsUsed": "0",
                "Capacity": "1",
                "TimezoneUsed": "GMT +09:00",
                "TimezoneIana": "Asia/Tokyo",
                "TimezoneWin": "Tokyo Standard Time",
                "startDateTime": f"{date_str} 08:30",
                "endDateTime": f"{date_str} 20:00",
                "locationName": "Iceberg",
                "locationCity": "Tokyo",
                "locationCountry": "JPN",
                "locationState": ""
            },
            "coworkingPropertyId": 0,
            "locationId": location_id,
            "Notes": {
                "locationAddress": "6-12-18 Jingumae",
                "locationCity": "Tokyo",
                "locationState": "",
                "locationCountry": "JPN",
                "locationName": "Iceberg"
            },
            "isUpdateBooking": False,
            "reservationId": "",
            "startTime": f"{date_str}T00:00:00+09:00",
            "endTime": f"{date_str}T23:59:59+09:00"
        }
        return self.do_request('post', url, data)

    def post_booking(self, date_str, reservable_id, location_id):
        url = 'https://members.wework.com/workplaceone/api/ext-booking/post-booking?APIen=0'
        quote = self.mxg_poll_quote(date_str, location_id, reservable_id)
        if not quote:
            return None

        data = {
            "reservableId": reservable_id,
            "type": 4,
            "creditsUsed": 0,
            "orderId": quote["uuid"],
            "ApplicationType": "WorkplaceOne",
            "PlatformType": "WEB",
            "TriggerCalenderEvent": True,
            "mailData": {
                "dayFormatted": datetime.strptime(date_str, "%Y-%m-%d").strftime("%A, %B %d"),
                "startTimeFormatted": "08:30 AM",
                "endTimeFormatted": "20:00 PM",
                "floorAddress": "",
                "locationAddress": "2-24-12 Shibuya",
                "creditsUsed": "0",
                "Capacity": "1",
                "TimezoneUsed": "GMT +09:00",
                "TimezoneIana": "Asia/Tokyo",
                "TimezoneWin": "Tokyo Standard Time",
                "startDateTime": f"{date_str} 08:30",
                "endDateTime": f"{date_str} 20:00",
                "locationName": "Shibuya Scramble Square",
                "locationCity": "Tokyo",
                "locationCountry": "JPN",
                "locationState": ""
            },
            "coworkingPropertyId": 0,
            "applicationType": "WorkplaceOne",
            "platformType": "WEB",
            "locationId": location_id,
            "Notes": {
                "locationAddress": "2-24-12 Shibuya",
                "locationCity": "Tokyo",
                "locationState": "",
                "locationCountry": "JPN",
                "locationName": "Shibuya Scramble Square"
            },
            "isUpdateBooking": False,
            "reservationId": "",
            "startTime": f"{date_str}T00:00:00+09:00",
            "endTime": f"{date_str}T23:59:59+09:00"
        }
        res = self.do_request('post', url, data)

        if res:
            return BookSpaceResponse(res)

        return None

    def get_locations_by_geo(self, city):
        url = f'https://members.wework.com/workplaceone/api/wework-yardi/ondemand/get-locations-by-geo?isAuthenticated=true&city={city}&isOnDemandUser=false&isWeb=true'
        params = {}
        
        response = self.do_request('get', url, params)
        if response:
            return LocationsByGeoResponse(response)
        return None

    def get_available_spaces(self, date_str, location_uuid):
        url = f'https://members.wework.com/workplaceone/api/spaces/get-spaces?locationUUIDs={','.join(location_uuid)}&closestCity=&userLatitude=35.6953443&userLongitude=139.7564755&boundnwLat=&boundnwLng=&boundseLat=&boundseLng=&type=0&offset=0&limit=50&roomTypeFilter=&date={date_str}&duration=30&locationOffset=%2B09%3A00&isWeb=true&capacity=0&endDate='

        response = self.do_request('get', url)
        if response:
            return SharedWorkspaceResponse(response)
        return None
    

def main():
    parser = argparse.ArgumentParser(description="WeWork Booking CLI")
    parser.add_argument('action', choices=['book', 'spaces', 'locations'], help="Action to perform: 'book', 'spaces', or 'locations'")
    parser.add_argument('date', help="Date in YYYY-MM-DD format")
    parser.add_argument('--location-uuid', help="Location ID for booking")
    parser.add_argument('--city', help="City name (required when action is 'geo')")
    parser.add_argument("--username", help="Username", required=True)
    parser.add_argument("--password", help="Password", required=True)

    args = parser.parse_args()

    auth = WeWorkAuth(args.username, args.password)
    result = auth.authenticate()

    ww = WeWork(result["token"], result["idToken"])

    if args.action == 'book':
        if not args.location_uuid:
            print("Error: --location-uuid) are required for booking.")
            sys.exit(1)

        
        # Parse the date argument
        dates = []
        if '~' in args.date:
            # Date range
            start_date, end_date = args.date.split('~')
            start_date = datetime.strptime(start_date.strip(), '%Y-%m-%d')
            end_date = datetime.strptime(end_date.strip(), '%Y-%m-%d')
            delta = end_date - start_date
            for i in range(delta.days + 1):
                dates.append((start_date + timedelta(days=i)).strftime('%Y-%m-%d'))
        elif ',' in args.date:
            # Comma-separated list
            dates = [date.strip() for date in args.date.split(',')]
        else:
            # Single date
            dates = [args.date.strip()]

        # do lookup for location_id
        for date in dates:
            print(f"Checking availability for {date}")
            spaces = ww.get_available_spaces(date, [args.location_uuid])
            
            if not spaces or not spaces.workspaces:
                print("Error: No spaces found, or not available for the given date.")
                sys.exit(1)
            
            if len(spaces.workspaces) > 1:
                print("Found multiple spaces: ")
                for space in spaces.workspaces:
                    print(f"Location: {space.location.name}")
                    print(f"Reservable ID: {space.uuid}")
                    print(f"Location ID: {space.location.uuid}")
                    print(f"Available: {space.seat.available}")
                    print("---")

                print("please specify a specific space to book")
                sys.exit(1)

            # do booking
            space = spaces.workspaces[0]

            print(f"Attempting to book: {space.location.name} for {date}")
            book_res = ww.post_booking(date, space.uuid, space.location.uuid)
            if book_res.booking_processing_status == "BookingSuccess":
                print(f"Booking successful!")

    elif args.action == 'locations':
        if not args.city:
            print("Error: --city is required for location lookup.")
            sys.exit(1)

        res = ww.get_locations_by_geo(args.city)

        for location in res.locations:
            print(f"Location: {location.name}")
            print(f"UUID: {location.uuid}")
            print(f"Latitude: {location.latitude}")
            print(f"Longitude: {location.longitude}")
            print("---")

    elif args.action == 'spaces':
        if not args.location_uuid and not args.city:
            print("Error: --location-uuid or --city is required for spaces lookup.")
            sys.exit(1)


        if args.city:
            res = ww.get_locations_by_geo(args.city)
            location_uuid = [location.uuid for location in res.locations]
        else:
            location_uuid = args.location_uuid.split(",")

        response = ww.get_available_spaces(args.date, location_uuid)

        if not response or len(response.workspaces) == 0:
            print("No spaces found, or not available for the given date.")
            sys.exit(1)

        for space in response.workspaces:
            print(f"Location: {space.location.name}")
            print(f"Reservable ID: {space.uuid}")
            print(f"Location ID: {space.location.uuid}")
            print(f"Available: {space.seat.available}")
            print("---")

if __name__ == "__main__":
    main()
