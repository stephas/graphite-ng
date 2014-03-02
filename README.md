# Graphite-ng

Experimental version of a new generation Graphite API server in Golang,
leveraging Go's efficient concurrency constructs.

Goals are: speed, ease of deployment & elegant code.

Furthermore, this rewrite allows to fundamentally redesign some specific
annoyances that can't easily be changed in Graphite.


# Backend stores

Graphite-ng supports multiple stores.
The ideal metrics storage can easily scale from a few metrics on my laptop in power-save mode
to millions of metrics on a highly loaded cluster; something that automatically balances the load,
assures HA and heals itself in case of disk of node failures.
Currently implemented:

* text (for reads only)
* influxdb (for reads, you have to do your own ingestion for now)

Other candidates that you could hack on if you're interested

* elasticsearch (there's a carbon-es dir with an exerimental carbon daemon)
* [kairosdb](https://code.google.com/p/kairosdb/)
* [whisper](https://github.com/graphite-project/whisper) (good for legacy data, discouraged otherwise)
* [ceres](https://github.com/graphite-project/ceres) (good for legacy data, discouraged otherwise)


# Omissions and limitations

 * Only the json output, not the png renderer. (because [client side
   rendering](https://github.com/vimeo/timeserieswidget/) is way better, it can give you interactive graphs)
 * No web UI (because there are plenty of graphite dashboards out there)
 * No events system ([anthracite](https://github.com/Dieterbe/anthracite/) is
   better than the very basic graphite events thing)
 * Currently only a small set of functions are supported. (see `data.go` and the `functions/` dir.)
 * No wildcards yet
 * Metric identifiers must at least contain 1 dot
 * From/until paramaters can only be in unix timestamp format. (luckily most dashboards abstract this away nicely)

# How it works

`graphite-ng` is a webserver that gives you a `/render/` http endpoint where
you can do queries like
`/render/?target=sum(test.metric1,scale(test.metric2,5.2))&from=123&until=456`

`graphite-ng` converts all user input into a real, functioning Go program,
compiles and runs it, and returns the output. It can do this because the
graphite api notation can easily be converted to real program code. Great
power, great responsability. The worker functions use goroutines and channels
to stream data around and avoid blocking.

For every metric, it will automatically find the right store in the order from the config file.
So you can have metrics in different stores and graphite-ng will automatically figure out what's where.

# Installation & running

Run this from the code checkout:

    rm -f executor-*.go ; go install github.com/graphite-ng/graphite-ng && graphite-ng

Then open something like this in your browser:

    http://localhost:8080/render/?target=test.metric2&target=derivative(test.metric1)
    http://localhost:8080/render/?target=sum(test.metric1,scale(test.metric2,5))&from=60&until=300

These test metrics are available by default through the text store.  You'll probably want
to add your real metrics into influxdb.  Which, by the way, is [really easy](http://influxdb.org/docs/) to install.


# Function plugins

All functions come in plugin files. want to add a new function? just drop a .go
file in the functions folder and restart. You can easily add your own functions
that get data from external sources, manipulate data, or represent data in a
different way; and then call those functions from your target string.


# Other interesting things & diff with real graphite

* Every function can request a different timerange from the functions it
  depends on. E.g.:
  * `derivative` needs the datapoint from before the requested timerange
  * `movingAverage(foo, X)` needs x previous datapoints. Regular graphite
	doesn't support this so you end up with gaps in the beginning of the graph.
* Clever automatic rollups based on tags (TODO)
* The `pathExpression` system in graphite is overly complicated and buggy. The
  keys in Graphite's json output are sometimes not exactly the requested target
  string (i.e. floats being rounded), it's not so easily fixed in Graphite
  which means client renderes have to implement ugly hacks to work around this.
  With graphite-ng we just use the exact same string.
* Use terminology from math and statistics properly, be correct and consistent.
* Avoid any results being dependent on any particular potentially unknown
  variable, aim for per second instead of per current-interval, etc.
  specifically:
  * `derivative` is a true derivative (ie `(y2-y1)/(x2-x1)`) unlike graphite's
	derivative where you depend on a factor that depends on whatever the
	resolution is at each point in time.
* Be mathematically/logically correct by default ("nil+123" should be "nil" not
  "123", though the functions could get a "sloppyness" argument)

# Community
  meet us in `#graphite-ng` on freenode
