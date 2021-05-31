User Guide
==========

## Installations
 
- Open a terminal and type the following command:
 
```
$ go get -v ./...
```
This should install all the non-standard packages needed for the project to work properly.
if the previous command display some error, please try the next one instead:
```
$ go get .
```
We'll be working with many terminals so have some ready to work.
# Server
We can start the server with the following command:
```
$ export GO111MODULE=off
$ go run main.go
```
With the server started we can use the other components of the project.

# Workers
In another terminal (one for each worker) create a worker with the following command:
```
$ go run worker/main.go --controller <controller adress> --worker-name <name of worker>
```
With the worker created you can finally do some stuff.

## Image Processing
With the help of the api.go script, we can send request to the server.
- Login

The *Login* command will receive an username and a password, and will generate a unique token required for the other commands.

```
curl -X POST -u username:password http://localhost:8080/login
```

Once its done the terminal will display the following message:

```
{
	"message": "Hi <username>, welcome to the DPIP System",
	"token" "<Token generated>"
}
```
- Logout

The *Logout* command will receive the token of the user and will revoke it.

```
$ curl -X DELETE -H "Authorization: <Token>" http://localhost:8080/logout
```

Once this is done the following message will appear:

```
{
	"message": "Bye <username>, your token has been revoked"
}
```

**Note: Because the token is revoked it will no longer be valid to use, please use this command at the very end**

- Create Workloads

the *workloads* command will create one workload. The job of these is to apply the filter to the images we will upload.

```
$ curl -X POST -H "Authorization: <Token>" -d '{"filter":"blur", "workload_name":""}' http://localhost:8080/workloads
```
If it creates the workload succesfully the following image will be displayed.
```
{
	Filter": "<Blur or Grayscale>",
	"Filtered Images": "<#>",
        "Message": "The workload has been successfully created",
        "Running Jobs": "<#>",
        "Status": "<Scheduling, Running, Completed>",
        "Workload ID": "<#>",
        "Workload Name": "<Name>"
}
```
- View Workloads

the *workloads/<workload_id>* command will display the information of the specified workload. 
```
$ curl -X GET -H "Authorization: <Token>" http://localhost:8080/workloads/<workload_id>
```
If it creates the workload succesfully the following image will be displayed.
```
{
     "Filter": "<Blur or Grayscale>",
     "Filtered Images": "<#>",
     "Message": "Information retrieved successfully",
     "Running Jobs": "<#>",
     "Status": "<Scheduling, Running, Completed>",
     "Workload ID": "<#>",
     "Workload Name": "<Name>"
}
```
- Status

the *Status* command will receive the token of the user.

```
$ curl -X -H "Authorization: Bearer <ACCESS_TOKEN>" http://localhost:8080/status
```

If the token exists it will the display the Number of active Workloads.

```
{
	"Active Workloads": "<#>",
     "Message": "Hi <username>, the DPIP System is Up and Running",
     "Time": "<Actual Time>"
}
```

- Upload Images

The *images* command will receive the **absolute** path of the image that you want to upload and the token of the user.

```
$ curl -X POST -F 'data=<Absolute Path>' -H "Authorization: <Token>" -d {"workload_id":<number of workload>} http://localhost:8080/images
```

If the Image is uploaded succesfully the following message will appear:

```
{
	"message": "An image has been successfully uploaded",
	"workload_id": "#",
	"image_id": "#",
	"type": "<image type>", 
}
```

- Download Image

The *images/<image_id>* command will Download the specified image.

```
$ curl -X GET -H "Authorization: <Token>" http://localhost:8080/images/<image_id>
```

If Everything is in order it will display the following message.

```
{
	"message": "An image has been Downloaded succesfully",
	"workload_id": "#",
	"image_id": "#",
	"type": "<image type>", 
}
```
