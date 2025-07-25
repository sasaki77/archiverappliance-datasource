
CREATE TABLE PVTypeInfo (
	pvName VARCHAR(255) NOT NULL PRIMARY KEY,
	typeInfoJSON MEDIUMTEXT NOT NULL,
	last_modified TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
) ENGINE=InnoDB;

CREATE TABLE PVAliases (
	pvName VARCHAR(255) NOT NULL PRIMARY KEY,
	realName VARCHAR(256) NOT NULL,
	last_modified TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
) ENGINE=InnoDB;

CREATE TABLE ArchivePVRequests (
	pvName VARCHAR(255) NOT NULL PRIMARY KEY,
	userParams MEDIUMTEXT NOT NULL,
	last_modified TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
) ENGINE=InnoDB;

CREATE TABLE ExternalDataServers (
	serverid VARCHAR(255) NOT NULL PRIMARY KEY,
	serverinfo MEDIUMTEXT NOT NULL,
	last_modified TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
) ENGINE=InnoDB;

