CREATE TABLE IF NOT EXISTS `documents` (
  `id` varchar(32) NOT NULL,
  `title` varchar(150) NOT NULL,
  `message` varchar(400) DEFAULT NULL,
  `status` varchar(45) NOT NULL,
  `void_reason` varchar(200) DEFAULT NULL,
  `kind` varchar(45) DEFAULT NULL,
  `demo` tinyint(1) NOT NULL DEFAULT '0',
  `created` datetime NOT NULL,
  `void_date` varchar(45) DEFAULT NULL,
  `pages` int(11) DEFAULT NULL,
  `user_id` varchar(32) NOT NULL,
  `enterprise_id` varchar(32) DEFAULT NULL,
  `s_message_id` varchar(100) DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=latin1;

CREATE TABLE IF NOT EXISTS `documents_security` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `document_id` varchar(250) NOT NULL,
  `sum` varchar(128) NOT NULL,
  `date` datetime NOT NULL,
  `event_id` varchar(100) DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=87 DEFAULT CHARSET=latin1;

CREATE TABLE IF NOT EXISTS `document_keys` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `document_id` varchar(250) NOT NULL,
  `master_key` varchar(250) DEFAULT NULL,
  `original_key` varchar(250) DEFAULT NULL,
  `certificate_key` varchar(250) DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=latin1;

CREATE TABLE IF NOT EXISTS `keys` (
  `key_id` varchar(100) NOT NULL,
  `secret_key` varchar(100) NOT NULL,
  `expiry` varchar(100) NOT NULL,
  `user_id` varchar(100) NOT NULL,
  `active` tinyint(1) DEFAULT NULL,
  PRIMARY KEY (`key_id`)
) ENGINE=InnoDB DEFAULT CHARSET=latin1;

CREATE TABLE IF NOT EXISTS `pages` (
  `id` varchar(150) NOT NULL,
  `document_id` varchar(150) NOT NULL,
  `last_updated` datetime NOT NULL,
  `bucket_key` varchar(150) NOT NULL,
  `order` int(11) NOT NULL,
  `width` decimal(10,2) NOT NULL,
  `height` decimal(10,2) NOT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=latin1;

CREATE TABLE IF NOT EXISTS `recipients` (
  `id` varchar(250) NOT NULL,
  `document_id` varchar(250) NOT NULL,
  `first_name` varchar(150) NOT NULL,
  `last_name` varchar(150) DEFAULT NULL,
  `status` varchar(400) NOT NULL,
  `email` varchar(250) NOT NULL,
  `routing` int(11) DEFAULT NULL,
  `created` datetime NOT NULL,
  `active` tinyint(1) NOT NULL DEFAULT '1',
  `next_reminder` datetime DEFAULT NULL,
  `complete` datetime DEFAULT NULL,
  `user_id` varchar(100) DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=latin1;

CREATE TABLE IF NOT EXISTS `sessions` (
  `id` varchar(32) NOT NULL,
  `recipient_id` varchar(32) NOT NULL COMMENT 'Denotes the recipient that the session has been created for.',
  `created` datetime NOT NULL COMMENT 'Denotes when the session was created.',
  `user_agent` varchar(250) NOT NULL COMMENT 'Denotes the software acting on behalf of the recipients HTTP call for the session.',
  `ip_address` varchar(45) NOT NULL COMMENT 'Ip address of the signer.',
  `geo_lat` decimal(10,8) DEFAULT NULL COMMENT 'Latitude of the geographic location of the signer',
  `geo_long` decimal(11,8) DEFAULT NULL COMMENT 'Longitude of the geographic location of the signer',
  `agreed` tinyint(1) DEFAULT NULL,
  `agreed_date` datetime DEFAULT NULL,
  `expiry` datetime NOT NULL COMMENT 'Denotes when the session will expire. After it has expired the recipient can no longer interface with the session.',
  `security` varchar(50) DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=latin1;


CREATE TABLE IF NOT EXISTS `tabs` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `x` decimal(10,2) NOT NULL,
  `y` decimal(10,2) NOT NULL,
  `page` varchar(100) NOT NULL,
  `kind` varchar(100) NOT NULL,
  `recipient_id` varchar(200) DEFAULT NULL,
  `size` varchar(45) DEFAULT NULL,
  `height` DECIMAL(10,2) DEFAULT NULL,
  `width` DECIMAL(10,2) DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=588 DEFAULT CHARSET=latin1;

CREATE TABLE IF NOT EXISTS `user` (
  `id` varchar(32) NOT NULL,
  `first_name` varchar(100) NOT NULL,
  `last_name` varchar(100) NOT NULL,
  `email` varchar(100) NOT NULL,
  `password` varchar(100) NOT NULL,
  `kind` int(11) NOT NULL DEFAULT '1',
  `next_kind` int(11) NOT NULL DEFAULT '0',
  `enterprise_id` varchar(32) DEFAULT NULL,
  `s_customer_id` varchar(50) DEFAULT NULL,
  `demo` tinyint(1) NOT NULL DEFAULT '0',
  `verification_token` varchar(250) DEFAULT NULL,
  `verification_token_secret` varchar(250) DEFAULT NULL,
  `verification_token_expiry` datetime DEFAULT NULL,
  `reset_token` varchar(250) DEFAULT NULL,
  `reset_token_secret` varchar(250) DEFAULT NULL,
  `reset_token_expiry` datetime DEFAULT NULL,
  `current_period_end` datetime DEFAULT NULL,
  `failed_payments` int(11) NOT NULL DEFAULT '0',
  `downgrade_date` datetime DEFAULT NULL,
  `created` datetime DEFAULT CURRENT_TIMESTAMP,
  `signup_method` varchar(50) DEFAULT NULL,
  `branding_id` varchar(32) DEFAULT NULL,
  `email_verified` tinyint(1) NOT NULL DEFAULT '0',
  `group_active` tinyint(1) NOT NULL DEFAULT '0',
  `active` tinyint(1) NOT NULL DEFAULT '0',
  PRIMARY KEY (`id`),
  UNIQUE KEY `email_UNIQUE` (`email`)
) ENGINE=InnoDB DEFAULT CHARSET=latin1;


CREATE TABLE IF NOT EXISTS `events` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `document_id` varchar(200) NOT NULL,
  `body` varchar(300) NOT NULL,
  `created` datetime NOT NULL,
  `kind` varchar(45) DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=6537 DEFAULT CHARSET=latin1;

CREATE TABLE IF NOT EXISTS `events_security` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `event_id` int(11) NOT NULL,
  `sum` varchar(128) NOT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=latin1;

CREATE TABLE IF NOT EXISTS `correspondences` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `recipient_id` varchar(200) NOT NULL,
  `method` varchar(100) NOT NULL,
  `sent` datetime NOT NULL,
  `third_party_id` varchar(200) NOT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=latin1;

CREATE TABLE IF NOT EXISTS `enterprises` (
  `id` varchar(32) NOT NULL,
  `name` varchar(250) NOT NULL,
  `address` varchar(250) NOT NULL,
  `contact` varchar(250) NOT NULL,
  `branding_id` varchar(32) DEFAULT NULL,
  `seats` int(11) NOT NULL DEFAULT '1',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=latin1;

CREATE TABLE `entrypoints` (
  `id` varchar(250) NOT NULL,
  `recipient_id` varchar(32) NOT NULL,
  `secret` varchar(250) NOT NULL,
  `auth` varchar(50) NOT NULL,
  `expiry` datetime DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=latin1;

CREATE TABLE IF NOT EXISTS `signatures` (
  `id` varchar(36) NOT NULL,
  `bucket_key` varchar(250) NOT NULL,
  `thumb_key` varchar(250) DEFAULT NULL,
  `thumb_height` int(4) DEFAULT NULL,
  `thumb_width` int(4) DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=latin1;

CREATE TABLE IF NOT EXISTS `user_signatures` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `signature_id` varchar(36) NOT NULL,
  `active` tinyint(1) NOT NULL DEFAULT '1',
  `user_id` varchar(32) NOT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=5 DEFAULT CHARSET=latin1;

CREATE TABLE IF NOT EXISTS `session_signatures` (
  `id` varchar(32) NOT NULL,
  `signature_id` varchar(36) NOT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=latin1;

CREATE TABLE IF NOT EXISTS `user_quota` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `user_id` varchar(32) NOT NULL,
  `quota_reset` datetime NOT NULL,
  `quota` int(11) NOT NULL DEFAULT '0',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=52 DEFAULT CHARSET=latin1;

CREATE TABLE IF NOT EXISTS `email_events` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `event_id` varchar(50) NOT NULL,
  `email` varchar(200) NOT NULL,
  `state` varchar(45) NOT NULL,
  `time` datetime NOT NULL,
  `user_agent` varchar(200) DEFAULT NULL,
  `bounce_description` varchar(200) DEFAULT NULL,
  `diag` varchar(200) DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

CREATE TABLE IF NOT EXISTS `email_location` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `event_id` varchar(50) NOT NULL,
  `country` varchar(100) DEFAULT NULL,
  `latitude` decimal(10,8) DEFAULT NULL,
  `longitude` decimal(11,8) DEFAULT NULL,
  `city` varchar(100) DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

CREATE TABLE IF NOT EXISTS `email_opens` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `event_id` varchar(50) NOT NULL,
  `time` datetime NOT NULL,
  `kind` varchar(10) NOT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

CREATE TABLE IF NOT EXISTS `email_smtp_events` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `event_id` varchar(50) NOT NULL,
  `destination_ip` varchar(45) DEFAULT NULL,
  `diag` varchar(200) DEFAULT NULL,
  `time` datetime DEFAULT NULL,
  `kind` varchar(50) DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

CREATE TABLE IF NOT EXISTS `ecomm_failures` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `event_id` varchar(50) NOT NULL,
  `created` datetime NOT NULL,
  `s_customer_id` varchar(50) NOT NULL,
  `kind` varchar(100) NOT NULL,
  `invoice` varchar(50) DEFAULT NULL,
  `failure_message` varchar(100) DEFAULT NULL,
  `failure_code` varchar(50) DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

CREATE TABLE IF NOT EXISTS `branding` (
  `id` varchar(32) NOT NULL,
  `logo_url` varchar(250) DEFAULT NULL,
  `colour_scheme` varchar(250) DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
