# Spotify Analytics

Hobby project for analysing and enhancing my Spotify usage. Since Spotify offers premium users a "free" API key, might as well make use of it! Simultaneously, I wanted to try out Go to see how it compared with other languages I'm more familiar with.

The project is delivered as an API piggybacking on Spotify's own auth service. 

## Run locally

Pre-reqs:
- Docker
- go
- ngrok

App runs on port 8080 by default.

1. Clone repo
2. `docker compose -f docker-compose.dev.yml up --build`
3. Create an integration in [Spotify's developer dashboard](https://developer.spotify.com/dashboard)
4. Create a free ngrok account
5. Open a cmd prompt and run `ngrok http 8080`. Take a note of the internet-facing URL ngrok gives you whilst this is running
6. Create a file `.env` in the project root using the example `.env.example`

### Postman

Suggest calling the API from Postman since this has in-built OAuth helpers. In the /postman directory is a sample Postman collection for calling the API.

1. Import collection into Postman
2. Create the following environment variables using the details of your spotify app-
    - `spotifyClientId`
    - `spotifyClientSecret`
3. If you changed the port of the app, alter the collection variable `baseUrl`
