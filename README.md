# Spiny Dogfish 

Spiny Dogfish is a Java Spring Properties Pruner.  Spiny Dogfish will reduce the clutter of Java Spring properties files making the actual configuration for your environment more apparent.  
This will also make it easier to make changes to a configuration without unintended side effects.

Spiny Dogfish will look at files on your application's classpath as well as properties within an external configuration directory.  
The directories to be scanned are specified in the config.toml file in the same working directory as the executable.

Configuration files will be pruned to respect the [load order of configuration files from Spring Boot.](https://docs.spring.io/spring-boot/docs/current/reference/html/spring-boot-features.html#boot-features-external-config)
For easy reference, the following order will be used by the pruner with the first item taking the highest precendence: 

1. Profile-specific application properties outside of your packaged jar (application-{profile}.properties and YAML variants).

2. Profile-specific application properties packaged inside your jar (application-{profile}.properties and YAML variants).

3. Application properties outside of your packaged jar (application.properties and YAML variants).

4. Application properties packaged inside your jar (application.properties and YAML variants).

Spiny Dogfish will load either yaml or java properties files.  The application will output yaml files as well as a changeset.

## Running The App
1. Download the appropriate binary for your platform.  The binaries can be found under the releases tab of github.
2. Create a file called 'config.toml' in the same directory as the binary file.  Set the root directory for the project.  See the config.toml file in the repo for an example file.  The value for 'project_root' must be set.  external_properties does not need to be set, but it should be blank if it will not be used.  Windows users should use forward slashes rather than backslashes, ie c:/my-dev-directory/project 
3. Run the application using ./<spiny-dogfish-executable> or <spiny-dogfish-executable>.exe 

## Known Issues

* Properties with camelcase keys will not be properly imported or exported.  This may result in duplicate key values and when the key name is exported it may not match the key used within your application for the property
* Properties with list values are not properly exported.   
* The command line in windows does not display correctly.

