import urllib.request

request = urllib.request.Request('https://api.github.com/user', headers={
    'Accept': 'application/vnd.github+json',
    'Authorization': f'Bearer {open("token.txt").read()}',
    'X-GitHub-Api-Version': '2022-11-28'
})

response = urllib.request.urlopen(request)

print(response.read())
