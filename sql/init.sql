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
-- Create "meter_items" table
CREATE TABLE `meter_items` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `created_at` timestamp NOT NULL,
  `updated_at` timestamp NOT NULL,
  `deleted_at` timestamp NULL,
  `code` varchar(255) NOT NULL,
  `unit` varchar(255) NOT NULL,
  `is_system` bool NOT NULL DEFAULT 0,
  `calc_type` enum('DIFFERENCE','DIRECT') NOT NULL DEFAULT "DIFFERENCE",
  `remark` varchar(255) NULL,
  `account_id` bigint NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE INDEX `meteritem_account_id_code` (`account_id`, `code`),
  CONSTRAINT `meter_items_accounts_meter_items` FOREIGN KEY (`account_id`) REFERENCES `accounts` (`id`) ON UPDATE NO ACTION ON DELETE NO ACTION
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
-- Create "meter_readings" table
CREATE TABLE `meter_readings` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `period_month` varchar(255) NOT NULL,
  `previous_reading` double NULL,
  `current_reading` double NOT NULL,
  `usage_amount` double NULL,
  `reading_time` timestamp NOT NULL,
  `reader_id` bigint NULL,
  `source` enum('MANUAL','IMPORT','AUTO') NOT NULL DEFAULT "MANUAL",
  `remark` varchar(255) NULL,
  `room_id` bigint NOT NULL,
  PRIMARY KEY (`id`),
  INDEX `meterreading_period_month` (`period_month`),
  UNIQUE INDEX `meterreading_room_id_period_month` (`room_id`, `period_month`),
  CONSTRAINT `meter_readings_rooms_meter_readings` FOREIGN KEY (`room_id`) REFERENCES `rooms` (`id`) ON UPDATE NO ACTION ON DELETE NO ACTION
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
-- Create "room_meter_items" table
CREATE TABLE `room_meter_items` (
  `room_id` bigint NOT NULL,
  `meter_item_id` bigint NOT NULL,
  PRIMARY KEY (`room_id`, `meter_item_id`),
  INDEX `room_meter_items_meter_item_id` (`meter_item_id`),
  CONSTRAINT `room_meter_items_meter_item_id` FOREIGN KEY (`meter_item_id`) REFERENCES `meter_items` (`id`) ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT `room_meter_items_room_id` FOREIGN KEY (`room_id`) REFERENCES `rooms` (`id`) ON UPDATE NO ACTION ON DELETE CASCADE
) CHARSET utf8mb4 COLLATE utf8mb4_bin;
