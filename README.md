# Mud Shark 

Mud Shark is a Java Spring Properties Pruner.  Mud Shark will reduce the clutter of Java Spring properties files making the actual configuration for your environment more apparent.  
This will also make it easier to make changes to a configuration without unintended side effects.

Mud Shark will look at files on your application's classpath as well as properties within an external configuration directory.  
The directories to be scanned are specified in the config.toml file in the same working directory as the executable.

Configuration files will be pruned to respect the load order of configuration files from Spring Boot.  
The following order will be used by the pruner: 

https://docs.spring.io/spring-boot/docs/current/reference/html/spring-boot-features.html#boot-features-external-config

4. Profile-specific application properties outside of your packaged jar (application-{profile}.properties and YAML variants).

3. Profile-specific application properties packaged inside your jar (application-{profile}.properties and YAML variants).

2. Application properties outside of your packaged jar (application.properties and YAML variants).

1. Application properties packaged inside your jar (application.properties and YAML variants).

Mud Shark will load either yaml or java properties files.  The application will output yaml files as well as a changeset.

