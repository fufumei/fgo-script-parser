## TODO

- Only output to a file if it's directory/batch parsing?
- Figure out how to deal with events and interludes. Right now, it only checks for "main". Events seem to also use "main" type, so that should be fine. Interludes use `friendship`.
  - Maybe use another flag for interlude and just use the quest API with the interlude quest ID? (group it and make it only available with "Quest" flag)
- Have a basic translation table set up for main quest war IDs and names, so the program always knows what to use (and can flag to use folder names if local)?
- Figure out a way to check multiple wars at once against atlas
- Run fetching and counting in goroutines for parallelism
- Verbose flag
- Cleaning needs to take into account the top line when using the raw script from atlas
