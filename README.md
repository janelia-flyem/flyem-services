# flyem-services

Implements authentication and authorization for a set of different
services.  Each service gets a unique and private JWT token.


## Installation

Go must be installed and GOPATH must be set to a location to store the spplication.

installation:

    % go get github.com/janelia-flyem/flyem-services

## Running

    % flyem-services -p |PORTNUM| config.json

See the example config.json.example file.  The user needs to register the application with
google oauth.  This file should specify the different services supported and their private
service secret.  The service secret will be used to create JWT specific to that application.  Once
the user is authenticated with flyem-services, the user can get tokens for any
appliation by accessing /api/token/APPNAME.  The token can be decoded by the application
specific secret.  The JWT contains the email address, profile picture, expiration time, and authorization information.


## Authorization

By default, the tokens generate will have permission level of "noauth".  A file
can be specified as in the example "authorizedvid.json".  By default, "noauth"
and "admin" should be considered protectd user authorization levels. Otherwise,
arbitrary objects can be stored here.  They will be packaged into the JWT.
If no authorization
is available for the user "noauth" will be used. 

## For service provides

* Create a password for your app and add to the config file
* Add authorization file if desired
* Add logic to decode JWT in your service with the service secret

## For clients

* Review swagger documentation at /help
* To login and get a token for application FOO: /login?redirect=/api/token/FOO
* To see all the supported applications: /api/applications
* Logout will require re-login for any service managed by this application: /logout

## Future Work

* Enable the service to use cloud functions for authorization storage
* Ability to launch the service as a cloud function

