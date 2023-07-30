# Download Google Web Fonts

A small Go executable to download Google web fonts for self hosting fonts on your own servers.

## Why?

Because self hosting is cool again and loading fonts from your own infrastructure means there is one less privacy concern which you have to list in your privacy policy.

## How?

Build the project using `go build`.

Then you can call the CLI using `./download-google-fonts`.

### Supported arguments:

| Arg | Description |
| --- | --- |
| `-url` | The URL of the Google Web Fonts to download |
| `-output` | The absolute or relative path of the desired output directory |
| `-destination` | An optional URL to a CDN or server location where to fonts will be hosted |

The `-destination` argument is only used to generate the self hosted URLs in the final CSS file. If this argument is not provided then the fonts will be referenced as if they live under the same path where the CSS file will be hosted.

### Examples:

```bash
./download-google-fonts -url "https://fonts.googleapis.com/css2?family=Lato:ital,wght@0,400&display=swap" -destination "https://cdn.my-server.com/fonts" -output "fonts"
```

The actual font files (e.g. `*.woeff2`) will be named with the MD5 hash of the file itself. This ensures that any updates to the font will generate a new file and allow updating fonts easily without running into caching issues.