# API server template

A template to setup an API project using Fiber Go web application framework with frontend artifacts built with vite

## Configuration

edit docker-compose.yml to match your database and redis settings

Configuration files are stired in a folder with a random name (change it)

See the WOIRTUMNSDFOEWR983745 folder.  it contains config files for 3 environments. Edit them to match your environment and host settings.

## Building and deploy

Building and deploying the API server is done via makefile.  The makefile has targets for building, testing, deploying, etc.

## secure access to API routes

In some cases you may not want the data to be accessible to anyone.  For example, you may not want to allow anyone to access the list of users in the system.

To secure access to API routes, use the VerifyFormSignature function in the services package.  This function checks the signature in the request body against the user's API key.  If the signature does not match, an error is returned and the request is not processed.


API calls from a mobile app and web UI via POST can include a signature in the request body.  The APi handlers for POST requests can then include this code to verify the signature;

```
		if _, err := services.VerifyFormSignature(db, c); err != nil {
			fmt.Println(err)
			c.Status(503).SendString(err.Error())
			return err
		}
```

## Development and Testing

In order to test API handlers from rest.http exclude this test if TEST_MODE or USE_DOCKER is defined as "true";
```
if os.Getenv("TEST_MODE") == "true" || database.GetParam("USE_DOCKER") == "true" {
			// no-op
} else {
		if _, err := services.VerifyFormSignature(db, c); err != nil {
			fmt.Println(err)
			c.Status(503).SendString(err.Error())
			return err
		} 
}
```

