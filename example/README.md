# Train on custom data, an example.
* Make sure you have docker installed and optionally nvidia-docker
  * If you wish to train on your GPU for increased performance. Then enable this in the `docker/Makefile` file, by setting `GPU=1`                                                                                         
* Run `docker build -t darknetw .` in the root of the project
* Run `docker run -it --entrypoint bash darknetw`
  * Add `--gpus=all` flag if you have `GPU=1`
* Then within the docker container run `./darknetw generate --output example/dataset --images 500` from the project root or use your own dataset.
    * This will generate a simple dataset with rectangles and circles
* Optionally create your own dataset validation and test split (a 90/10 split is usually ok). Not required as valid.txt and train.txt is already provided.
```
find dataset | grep jpeg$ | shuf > /tmp/files.list
cat /tmp/files.list | head -n $(expr `cat /tmp/files.list | wc -l` / 10) > valid.txt
cat /tmp/files.list | grep -v -f valid.txt > train.txt
```


* (Optional) Create dataset config (dataset.cfg).
```
classes = 2
train  = train.txt
valid  = valid.txt
names = names.txt
```
* (Optional) Create names.txt (this holds the class names)
```
circle
rectangle
```

* Set correct number of classes in yolo config file (network.cfg) (https://github.com/AlexeyAB/darknet#how-to-train-to-detect-your-custom-objects), this is optional unless you are using your own dataset with a different number of classes.

* Run `./darknetw train` with the appropriate command line flags or environment variables or run `./darknetw serve` to spin up the
webservice and invoke the `POST /api/v1/train` endpoint.

GLHF
