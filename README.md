# dir2cast

Turn .mp3 files in current directory to a podcast feed just one command. Then you can subscribe to it with your favorite podcast client, download it for offline listening.

```
$ ls
ep1.mp3 ep2.mp3
$ dir2cast
[Parse] Processing ep1.mp3
[Parse] Processing ep2.mp3
[HTTP] Listening on port 8080 ...
```

1. Open your podcast client.
2. Choose `Subscribe by URL` and paste the URL: `http://<your LAN address>:8080/feed.xml`.
3. Download and cache all episodes.
4. Now you can stop the `dir2cast` proccess.

`dir2cast` will use filename as episode titles and use file modification time as episode publish date, extract ID3 tags from mp3 files, like lyrics and album cover.

`dir2cast` will use current directory as the name of podcast, and the first album cover as the podcast icon.
