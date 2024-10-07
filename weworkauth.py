import requests
import json
import re
from urllib.parse import urlencode, parse_qs, urljoin
from bs4 import BeautifulSoup


class WeWorkAuth:
    def __init__(self, username, password):
        self.username = username
        self.password = password
        self.session = requests.Session()
        self.config = self.get_auth0_config()

    def get_auth0_config(self):
        url = "https://members.wework.com/workplaceone/api/auth0/config"
        params = {
            "companyId": "00000000-0000-0000-0000-000000000000",
            "domain": "members.wework.com"
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
            "code_challenge": "_HPGV4QLcSdpVDba3O_N-teol5pwfyeQt2-VQ0ZbeLo",
            "code_challenge_method": "S256",
            "auth0Client": "eyJuYW1lIjoiQGF1dGgwL2F1dGgwLWFuZ3VsYXIiLCJ2ZXJzaW9uIjoiMS4xMS4xIiwiZW52Ijp7ImFuZ3VsYXIvY29yZSI6IjEzLjEuMSJ9fQ=="
        }
        auth_url = f"https://{self.config['domain']}/authorize?{urlencode(auth_params)}"
        response = self.session.get(auth_url, allow_redirects=False)
        login_url = response.headers['Location']
        state = parse_qs(login_url.split('?')[1])['state'][0]

        # Step 2: Perform login
        login_url = f"https://{self.config['domain']}/usernamepassword/login"
        login_data = {
            "client_id": self.config["client_id"],
            "redirect_uri": self.config["redirect_uri"],
            "tenant": "wework-prod",
            "response_type": "code",
            "scope": "openid profile email offline_access",
            "audience": self.config["audience"],
            "state": state,
            "nonce": "ZURxb08tUXJtUVlDVlFPfmFwcm1za1h6Q2ZySmtZSUNGWXJJS3ItaWktbg==",
            "connection": "id-wework",
            "username": self.username,
            "password": self.password,
            "_csrf": "kefoaTyE-K6YNfddtoRMYnMJJH_KX6zInTy4",
            "_intstate": "deprecated",
            "protocol": "oauth2"
        }
        response = self.session.post(login_url, json=login_data)
        
        if response.status_code != 200:
            raise Exception(f"Login failed: {response.text}")

        # Step 3: Handle login response and extract form data
        soup = BeautifulSoup(response.text, 'html.parser')
        form = soup.find('form')
        if not form:
            raise Exception("Couldn't find form in response")

        form_action = form.get('action')
        form_data = {}
        for input_tag in form.find_all('input'):
            name = input_tag.get('name')
            value = input_tag.get('value')
            if name and value:
                form_data[name] = value

        # Step 4: Submit the form and handle redirects
        response = self.session.post(form_action, data=form_data, allow_redirects=False)
        
        while response.status_code in (301, 302, 303, 307, 308):
            redirect_url = response.headers['Location']
            # print(f"Redirecting to: {redirect_url}")
            
            # Handle relative URLs
            if not redirect_url.startswith(('http://', 'https://')):
                redirect_url = urljoin(f"https://{self.config['domain']}", redirect_url)
            
            if 'code' in parse_qs(redirect_url.split('?', 1)[-1]):
                # We've reached the final redirect with the code
                code = parse_qs(redirect_url.split('?', 1)[-1])['code'][0]
                break
            
            if 'code' in parse_qs(redirect_url.split('?', 1)[-1]):
                # We've reached the final redirect with the code
                code = parse_qs(redirect_url.split('?', 1)[-1])['code'][0]
                break
            
            response = self.session.get(redirect_url, allow_redirects=False)
        else:
            raise Exception("Didn't receive expected code in redirects")

        # print(f"Obtained code: {code}")

        # Step 5: Exchange code for tokens
        token_url = f"https://{self.config['domain']}/oauth/token"
        token_data = {
            "client_id": self.config["client_id"],
            "code_verifier": "vLbUww3TSb7sJcJhtKuFbeVDWKdmkpamtqrhEjuOx.Y",
            "grant_type": "authorization_code",
            "code": code,
            "redirect_uri": self.config["redirect_uri"]
        }
        response = self.session.post(token_url, json=token_data)
        tokens = response.json()

        # Step 6: Login to WeWork backend
        return self.login_to_wework(tokens['id_token'])

    def login_to_wework(self, id_token):
        login_url = "https://members.wework.com/workplaceone/api/auth0/login-by-auth0-token"
        login_data = {"id_token": id_token}
        response = self.session.post(login_url, json=login_data)
        return response.json()

# Usage
# username = "sunny.gate0653@d.sh"
# password = "bqa9cea9ywk*wpv1XV"
# auth = WeWorkAuth(username, password)
# result = auth.authenticate()
# print(json.dumps(result, indent=2))