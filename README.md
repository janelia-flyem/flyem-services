# flyem-services

Implement authentication and authorization for a set of different
services.  Each service gets a unique and private JWT token.


## Installation

Go must be installed and GOPATH must be set to a location to store the spplication.

installation:

    % go get github.com/janelia-flyem/flyem-services

## Running

    % flyem-services -p |PORTNUM| config.json

See the example config.json file.  The user needs to register the application with
google oauth.  This file should specify the different services supported and their private
key.  The private will be used to create JWT specific to that application.  One
the user is authenticate with flyem-services, the user can get tokens for any
appliation by accessing /api/token/APPNAME.  The token can be decoded by the application
specific secret.


## Authorization

By default, the tokens generate will have permission level of "noauth".  A file
can be specified as in the example "authorizedvid.json".  By default, "noauth"
and "admin" should be considered protectd user authorization levels.  Otherwise,
arbitrary objects can be stored here.  They will be packaged into the JWT.

## Future Woek

* Enable the service to use cloud functions for authorization storage
* Ability to launch the service as a cloud function

