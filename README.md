# firecall

A simple api endpoint for registering firebase devices and then sending call notifications to them.

## Why

I was annoyed to always type phone numbers from my computer to my mobile phone (where i got SIP configured to call customers using my company number).

My solution was to build this simple service for playing the middleman with Firebase and then send notifications to my phone which get received by a small app then dialing every incoming phone number (the app will get up here once i made it a little less embarrassing).

## How

This service offers two endpoints:

* `/dial/register`: which accepts the devices messaging token (only using those for now) and thn returns your api key
* `/dial/call/{number}`: which takes the number to dial as stated in the path and a Authorization Header containing the api key.

## Development

This project is built using [Bazel](https://bazel.build).
To build and run the code directly, install Bazel and run `make run CMD=//cmd/firecall`.

We build directly into a docker image and deploy it to Kubernetes.
The image step can be triggered manually by running `make docker CMD=//cmd/firecall`.

To deploy directly to Kubernetes, run `make kube CMD=//cmd/firecall:dev` which we use internally as dev deployment target.
There will be a `:prod` target in the future.

## Coding and Style

Our code is always checked by Travis using `make test check` therefor all Golang rules on syntax and formating have to be met for pull requests to be merged.
While this might incur more work for possible contributors, we see the code produced here as production critical once finished and therefor strive for high code quality.

We are developing this mostly using TDD and BDD. If you don't know what this is, we recommend this [video](https://www.youtube.com/watch?v=uFXfTXSSt4I) for starters.

Please do reasonable commit sizes.

## Dependencies

All dependencies inside this project are being managed by [dep](https://github.com/golang/dep) and are checked in.
After pulling the repository, it should not be required to do any further preparations aside from `make deps` to prepare the dev tools (once).

If new dependencies get added while coding, make sure to add them using `dep ensure --add "importpath"` and to check them into git.
We recommend adding your vendor changes in a separate commit to make reviewing your changes easier and faster.

## Testing

To run tests you can use:

```bash
make test
```

## Contributing

Feedback and contributions are highly welcome. Feel free to file issues, feature or pull requests.
If you are interested in using this project now or in a later stage, feel free to get in touch.

## Attributions

* [Kolide for providing `kit`](https://github.com/kolide/kit)

## License

This project's license is located in the [LICENSE file](LICENSE).