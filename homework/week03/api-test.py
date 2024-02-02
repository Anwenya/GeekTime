import requests

session = requests.session()


def signup():
    url = "http://webook:31629/users/signup"
    jsons = {
        "email": "173777777771@qq.com",
        "password": "asjh123A&&",
        "confirmPassword": "asjh123A&&"
    }
    response = session.post(url, json=jsons)
    print(response.headers)
    print(response.status_code)
    print(response.text)


def login():
    url = "http://webook:31629/users/login"
    jsons = {
        "email": "173777777771@qq.com",
        "password": "asjh123A&&",
    }
    response = session.post(url, json=jsons)
    print(response.headers)
    print(response.status_code)
    print(response.text)


def profile():
    url = "http://webook:31629/users/profile"
    response = session.get(url)
    print(response.headers)
    print(response.status_code)
    print(response.text)


if __name__ == '__main__':
    signup()
    print("------------------")
    login()
    print("------------------")
    profile()
