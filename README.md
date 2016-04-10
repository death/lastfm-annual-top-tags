# Last.fm annual top tags

Compile top-tags-per-annum statistics from last.fm history.

# Example

    lastfm-annual-top-tags --api-key <api-key> --api-secret <api-secret> --user <user> --save my.json

Will grab the user's history and compile the statistics. It will also
save the raw statistical data to `my.json`.

Then you can load it and compile the statistics again:

    lastfm-annual-top-tags --load my.json

Example output:

    2012: doom metal (14.4), thrash metal (11.8), heavy metal (11.8), gothic metal (11.1), black metal (10.5)
    2013: doom metal (19.2), classic rock (8.8), neofolk (7.1), gothic metal (6.9), black metal (6.9)
    2014: classic rock (22.0), blues (14.7), doom metal (7.1), rock (6.0), punk (3.9)
    2015: classic rock (13.1), folk (12.7), post-punk (9.6), doom metal (5.9), rock (4.1)
    2016: post-punk (11.6), darkwave (9.3), synthpop (7.7), ebm (7.4), classic rock (5.8)

# License

MIT
