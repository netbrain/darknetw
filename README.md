# darknetw (darknet wrapper)
This is a opinionated wrapper around AlexeyAB's darknet fork with the intention of facilitating easy training, labelling and prediction through a REST API.

Features:

* Provides the following API endpoints
  * `POST /api/v1/predict`
  * `POST /api/v1/label`
  * `POST /api/v1/train`
  * `GET /api/v1/train`
  * `GET /api/v1/accuracy`
  * `DELETE /api/v1/accuracy`
* Organizes training sessions by storing a snapshot of dataset (hardlinked) and configuration upon training.
* Exposes the following commands through the `darknetw` executable
  * serve (starts the darknetw API service)
  * train (trains the neural network - equivalent of `darknet detector train`)
  * validate (validates the accuracy of the neural network - equivalent of `darknet detector map`)
  * generate (will create a simple computer generated test dataset with circles and rectangles in a random fashion)
* Available as a docker container

See the `example` folder to get started with training on custom data.

TODOs:

* [ ] Document API endpoints
* [ ] Document docker container
* [ ] Test the thing
