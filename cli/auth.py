import requests
import json
import re
from urllib.parse import urlencode, parse_qs, urljoin
from bs4 import BeautifulSoup

import pprint

import secrets
import hashlib
import base64


class WeWorkAuth:
    def __init__(self, username, password):
        self.username = username
        self.password = password
        self.session = requests.Session()
        self.config = self.get_auth0_config()
        # Generate PKCE values on init
        self.code_verifier = self._generate_code_verifier()
        self.code_challenge = self._generate_code_challenge()

    def _generate_code_verifier(self):
        # Generate a random code verifier
        token = secrets.token_urlsafe(32)
        return token

    def _generate_code_challenge(self):
        # Create code challenge using S256 method
        code_verifier_bytes = self.code_verifier.encode("ascii")
        digest = hashlib.sha256(code_verifier_bytes).digest()
        return base64.urlsafe_b64encode(digest).decode("ascii").rstrip("=")

    def get_auth0_config(self):
        url = "https://members.wework.com/workplaceone/api/auth0/config"
        params = {
            "companyId": "00000000-0000-0000-0000-000000000000",
            "domain": "members.wework.com",
        }
        response = self.session.get(url, params=params)
        return response.json()

    def authenticate(self):
        # Step 1: Get the initial state
        auth_params = {
            "redirect_uri": self.config["redirect_uri"],
            "client_id": self.config["client_id"],
            "audience": self.config["audience"],
            "scope": "openid profile email offline_access",
            "response_type": "code",
            "response_mode": "query",
            "nonce": "ZURxb08tUXJtUVlDVlFPfmFwcm1za1h6Q2ZySmtZSUNGWXJJS3ItaWktbg==",
            "code_challenge": self.code_challenge,
            "code_challenge_method": "S256",
            "auth0Client": "eyJuYW1lIjoiQGF1dGgwL2F1dGgwLWFuZ3VsYXIiLCJ2ZXJzaW9uIjoiMS4xMS4xLmN1c3RvbSIsImVudiI6eyJhbmd1bGFyL2NvcmUiOiIxMy4xLjEifX0=",
        }

        # print(auth_params)

        auth_url = f"https://{self.config['domain']}/authorize?{urlencode(auth_params)}"
        response = self.session.get(auth_url, allow_redirects=False)
        login_url = response.headers["Location"]
        state = parse_qs(login_url.split("?")[1])["state"][0]

        # print("Login URL:")
        # print(login_url)

        # Handle IDP redirect to /login
        idp_response = self.session.get(
            f"https://{self.config['domain']}{login_url}", allow_redirects=False
        )

        if idp_response.status_code in (301, 302, 303, 307, 308):
            login_url = idp_response.headers["Location"]

        # print("Login URL:")
        # print(login_url)
        # print(state)

        # Step 2: Perform actual login
        login_url = f"https://{self.config['domain']}/usernamepassword/login"
        login_data = {
            "client_id": self.config["client_id"],
            "redirect_uri": self.config["redirect_uri"],
            "tenant": "wework-prod",
            "response_type": "code",
            "scope": "openid profile email offline_access",
            "audience": self.config["audience"],
            "state": state,
            "nonce": "YlNIdnY3dWRJUW9zYlpKclNzamc4VDB4dkpmYUtrY3M4R2hLSUtnVW5JeQ==",
            "connection": "id-wework",
            "username": self.username,
            "password": self.password,
            "_csrf": "YDBviN84-fwB6GPKLCHbSHM_5hbfP6PIRlEw",
            "_intstate": "deprecated",
            "protocol": "oauth2",
            "popup_options": {},
            "sso": True,
            "response_mode": "query",
            "prompt": "login",
            "ui_locales": "en-US",
            "code_challenge_method": "S256",
            "code_challenge": self.code_challenge,
            "maintenance_mode_message": "Our systems are currently undergoing planned maintenance. While functionality is limited, your current bookings and access to WeWork buildings are not impacted. If you need additional assistance, please open a support ticket or contact us at support@wework.com.",
            "auth0Client": "eyJuYW1lIjoiQGF1dGgwL2F1dGgwLWFuZ3VsYXIiLCJ2ZXJzaW9uIjoiMS4xMS4xIiwiZW52Ijp7ImFuZ3VsYXIvY29yZSI6IjEzLjEuMSJ9fQ==",
            "enable_maintenance_mode": "false",
        }

        headers = {
            "Auth0-Client": "eyJuYW1lIjoiQGF1dGgwL2F1dGgwLWFuZ3VsYXIiLCJ2ZXJzaW9uIjoiMS4xMS4xIiwiZW52Ijp7ImFuZ3VsYXIvY29yZSI6IjEzLjEuMSIsImxvY2suanMtdWxwIjoiMTIuMC4yIiwiYXV0aDAuanMtdWxwIjoiOS4yMC4yIn19",
            "Content-Type": "application/json",
            "Origin": "https://idp.wework.com",
            "Referer": login_url,
        }

        # print("Login URL:")
        # print(login_url)

        # print("Headers:")
        # pprint.pprint(headers)

        # print("Login data:")
        # pprint.pprint(login_data)

        response = self.session.post(login_url, json=login_data, headers=headers)

        if response.status_code != 200:
            raise Exception(f"Login failed: {response.text}")

        # Step 3: Handle login response and extract form data
        soup = BeautifulSoup(response.text, "html.parser")
        form = soup.find("form")
        if not form:
            raise Exception("Couldn't find form in response")

        form_action = form.get("action")
        form_data = {}
        for input_tag in form.find_all("input"):
            name = input_tag.get("name")
            value = input_tag.get("value")
            if name and value:
                form_data[name] = value

        # Step 4: Submit the form and handle redirects
        response = self.session.post(form_action, data=form_data, allow_redirects=False)

        while response.status_code in (301, 302, 303, 307, 308):
            redirect_url = response.headers["Location"]
            # print(f"Redirecting to: {redirect_url}")

            # Handle relative URLs
            if not redirect_url.startswith(("http://", "https://")):
                redirect_url = urljoin(f"https://{self.config['domain']}", redirect_url)

            if "code" in parse_qs(redirect_url.split("?", 1)[-1]):
                # We've reached the final redirect with the code
                code = parse_qs(redirect_url.split("?", 1)[-1])["code"][0]
                break

            response = self.session.get(redirect_url, allow_redirects=False)
        else:
            raise Exception("Didn't receive expected code in redirects")

        # print(f"Obtained code, attempting token_exchange: {code}")

        # Step 5: Exchange code for tokens
        token_url = f"https://{self.config['domain']}/oauth/token"
        token_data = {
            "client_id": self.config["client_id"],
            "code_verifier": self.code_verifier,
            "grant_type": "authorization_code",
            "code": code,
            "redirect_uri": self.config["redirect_uri"],
        }
        response = self.session.post(token_url, json=token_data)
        tokens = response.json()

        # print("Tokens:")
        # pprint.pprint(tokens)

        # Step 6: Login to WeWork backend
        # print("Logging in to WeeWork backend...")
        return self.login_to_wework(tokens)

    def login_to_wework(self, tokens):
        login_url = (
            "https://members.wework.com/workplaceone/api/auth0/login-by-auth0-token"
        )
        # print("Requesting: " + login_url)
        # print("Body:")
        # pprint.pprint(tokens)

        headers = {
            "Host": "members.wework.com",
            "Request-Source": "com.wework.ondemand/WorkplaceOne/Prod/iOS/2.68.0(18.2)",
            "Accept": "application/json, text/plain, */*",
            "Sec-Fetch-Site": "same-origin",
            "Accept-Language": "en-US,en;q=0.9",
            "Accept-Encoding": "gzip, deflate, br",
            "Sec-Fetch-Mode": "cors",
            "Content-Type": "application/json",
            "Origin": "https://members.wework.com",
            "User-Agent": "Mobile Safari 16.1",
            "Referer": "https://members.wework.com/workplaceone/content2/login/authenticate/app-auth",
            "Connection": "keep-alive",
            "Sec-Fetch-Dest": "empty",
        }

        login_data = {
            "id_token": tokens["id_token"],
            "access_token": tokens["access_token"],
            "refresh_token": tokens["refresh_token"],
            "expires_in": tokens["expires_in"],
            "scope": tokens["scope"],
            "token_type": tokens["token_type"],
            "client_id": self.config["client_id"],
            "audience": self.config["audience"],
        }
        # print("Login Data")
        # pprint.pprint(login_data)
        response = self.session.post(login_url, json=login_data, headers=headers)
        return response.json()
