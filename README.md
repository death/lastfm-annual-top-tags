# Last.fm annual top tags

Compile top-tags-per-annum statistics from last.fm history.

For each track in a user's last.fm history, grabs the top tag for its
artist if needed, then increments a counter for this tag (in the
track's playing year). Then displays the top tags for each year.

# Example

    lastfm-annual-top-tags --api-key <api-key> --api-secret <api-secret> --user <user>

Will grab the user's history and compile the statistics. It will also
save the raw statistical data to `state.json`. If the file exist, will
update the statistics with the tracks played since last run.

Example output:

    2012: doom metal (14.4), thrash metal (11.8), heavy metal (11.8), gothic metal (11.1), black metal (10.5)
    2013: doom metal (19.2), classic rock (8.8), neofolk (7.1), gothic metal (6.9), black metal (6.9)
    2014: classic rock (22.0), blues (14.7), doom metal (7.1), rock (6.0), punk (3.9)
    2015: classic rock (13.1), folk (12.7), post-punk (9.6), doom metal (5.9), rock (4.1)
    2016: post-punk (9.1), doom metal (7.9), electronic (7.0), darkwave (6.8), neofolk (6.1)

# License

MIT
