-- Create "accounts" table
CREATE TABLE `accounts` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `created_at` timestamp NOT NULL,
  `updated_at` timestamp NOT NULL,
  `deleted_at` timestamp NULL,
  `name` varchar(255) NOT NULL,
  `email` varchar(255) NOT NULL,
  `password` varchar(255) NOT NULL,
  `is_active` bool NOT NULL DEFAULT 1,
  PRIMARY KEY (`id`)
) CHARSET utf8mb4 COLLATE utf8mb4_bin;
-- Create "buildings" table
CREATE TABLE `buildings` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `created_at` timestamp NOT NULL,
  `updated_at` timestamp NOT NULL,
  `deleted_at` timestamp NULL,
  `name` varchar(255) NOT NULL,
  `address` varchar(255) NOT NULL,
  `city` varchar(255) NOT NULL,
  `province` varchar(255) NOT NULL,
  `account_id` bigint NOT NULL,
  PRIMARY KEY (`id`),
  INDEX `buildings_accounts_buildings` (`account_id`),
  CONSTRAINT `buildings_accounts_buildings` FOREIGN KEY (`account_id`) REFERENCES `accounts` (`id`) ON UPDATE NO ACTION ON DELETE NO ACTION
) CHARSET utf8mb4 COLLATE utf8mb4_bin;
-- Create "permissions" table
CREATE TABLE `permissions` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `created_at` timestamp NOT NULL,
  `updated_at` timestamp NOT NULL,
  `deleted_at` timestamp NULL,
  `code` varchar(255) NOT NULL,
  `api_path` varchar(255) NOT NULL,
  `description` varchar(255) NOT NULL,
  `is_active` bool NOT NULL DEFAULT 1,
  PRIMARY KEY (`id`)
) CHARSET utf8mb4 COLLATE utf8mb4_bin;
-- Create "roles" table
CREATE TABLE `roles` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `created_at` timestamp NOT NULL,
  `updated_at` timestamp NOT NULL,
  `deleted_at` timestamp NULL,
  `name` varchar(255) NOT NULL,
  `account_id` bigint NOT NULL,
  `is_active` bool NOT NULL DEFAULT 1,
  PRIMARY KEY (`id`)
) CHARSET utf8mb4 COLLATE utf8mb4_bin;
-- Create "role_permissions" table
CREATE TABLE `role_permissions` (
  `role_id` bigint NOT NULL,
  `permission_id` bigint NOT NULL,
  PRIMARY KEY (`role_id`, `permission_id`),
  INDEX `role_permissions_permission_id` (`permission_id`),
  CONSTRAINT `role_permissions_permission_id` FOREIGN KEY (`permission_id`) REFERENCES `permissions` (`id`) ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT `role_permissions_role_id` FOREIGN KEY (`role_id`) REFERENCES `roles` (`id`) ON UPDATE NO ACTION ON DELETE CASCADE
) CHARSET utf8mb4 COLLATE utf8mb4_bin;
-- Create "room_types" table
CREATE TABLE `room_types` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `created_at` timestamp NOT NULL,
  `updated_at` timestamp NOT NULL,
  `deleted_at` timestamp NULL,
  `name` varchar(12) NOT NULL,
  `account_id` bigint NOT NULL,
  PRIMARY KEY (`id`),
  INDEX `room_types_accounts_room_types` (`account_id`),
  CONSTRAINT `room_types_accounts_room_types` FOREIGN KEY (`account_id`) REFERENCES `accounts` (`id`) ON UPDATE NO ACTION ON DELETE NO ACTION
) CHARSET utf8mb4 COLLATE utf8mb4_bin;
-- Create "rooms" table
CREATE TABLE `rooms` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `created_at` timestamp NOT NULL,
  `updated_at` timestamp NOT NULL,
  `deleted_at` timestamp NULL,
  `room_no` varchar(255) NOT NULL,
  `floor` bigint NOT NULL,
  `price` bigint NOT NULL,
  `size` bigint NOT NULL,
  `is_active` bool NOT NULL DEFAULT 1,
  `building_id` bigint NOT NULL,
  `room_type_id` bigint NOT NULL,
  PRIMARY KEY (`id`),
  INDEX `rooms_buildings_rooms` (`building_id`),
  INDEX `rooms_room_types_rooms` (`room_type_id`),
  CONSTRAINT `rooms_buildings_rooms` FOREIGN KEY (`building_id`) REFERENCES `buildings` (`id`) ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT `rooms_room_types_rooms` FOREIGN KEY (`room_type_id`) REFERENCES `room_types` (`id`) ON UPDATE NO ACTION ON DELETE NO ACTION
) CHARSET utf8mb4 COLLATE utf8mb4_bin;
