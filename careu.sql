-- phpMyAdmin SQL Dump
-- version 5.2.1
-- https://www.phpmyadmin.net/
--
-- Host: 127.0.0.1
-- Generation Time: Aug 01, 2026 at 08:06 AM
-- Server version: 10.4.32-MariaDB
-- PHP Version: 8.2.12

SET SQL_MODE = "NO_AUTO_VALUE_ON_ZERO";
START TRANSACTION;
SET time_zone = "+00:00";


/*!40101 SET @OLD_CHARACTER_SET_CLIENT=@@CHARACTER_SET_CLIENT */;
/*!40101 SET @OLD_CHARACTER_SET_RESULTS=@@CHARACTER_SET_RESULTS */;
/*!40101 SET @OLD_COLLATION_CONNECTION=@@COLLATION_CONNECTION */;
/*!40101 SET NAMES utf8mb4 */;

--
-- Database: `careu`
--

-- --------------------------------------------------------

--
-- Table structure for table `app_states`
--

CREATE TABLE `app_states` (
  `user_id` bigint(20) UNSIGNED NOT NULL,
  `state_json` longtext CHARACTER SET utf8mb4 COLLATE utf8mb4_bin NOT NULL CHECK (json_valid(`state_json`)),
  `updated_at` timestamp NOT NULL DEFAULT current_timestamp() ON UPDATE current_timestamp()
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

--
-- Dumping data for table `app_states`
--

INSERT INTO `app_states` (`user_id`, `state_json`, `updated_at`) VALUES
(2, '{\"patientProfile\":{\"name\":{\"en\":\"KFC grandpa\",\"bm\":\"KFC grandpa\",\"zh\":\"KFC grandpa\"},\"age\":100,\"gender\":\"male\",\"village\":\"Kampung Bunuk\",\"city\":{\"en\":\"Kuching\",\"bm\":\"Kuching\",\"zh\":\"Kuching\"},\"phone\":\"+60199894341\",\"email\":\"admin@gmail.com\"},\"history\":[{\"date\":\"2026-07-26\",\"systolic\":128,\"diastolic\":82,\"glucose\":118,\"hr\":76,\"sleep\":7.2,\"med\":true,\"symptoms\":[\"none\"]},{\"date\":\"2026-07-27\",\"systolic\":131,\"diastolic\":84,\"glucose\":126,\"hr\":78,\"sleep\":6.8,\"med\":true,\"symptoms\":[\"none\"]},{\"date\":\"2026-07-28\",\"systolic\":126,\"diastolic\":80,\"glucose\":198,\"hr\":80,\"sleep\":5.4,\"med\":false,\"symptoms\":[\"fatigue\"]},{\"date\":\"2026-07-29\",\"systolic\":138,\"diastolic\":88,\"glucose\":132,\"hr\":82,\"sleep\":6.1,\"med\":true,\"symptoms\":[\"none\"]},{\"date\":\"2026-07-30\",\"systolic\":149,\"diastolic\":93,\"glucose\":141,\"hr\":85,\"sleep\":4.9,\"med\":false,\"symptoms\":[\"dizziness\"]},{\"date\":\"2026-07-31\",\"systolic\":156,\"diastolic\":97,\"glucose\":150,\"hr\":88,\"sleep\":5.2,\"med\":true,\"symptoms\":[\"dizziness\",\"fatigue\"]},{\"date\":\"2026-08-01\",\"systolic\":120,\"diastolic\":80,\"glucose\":120,\"hr\":90,\"sleep\":9,\"med\":true,\"symptoms\":[\"none\"]}],\"medicationReminder\":{\"enabled\":true,\"time\":\"08:00\",\"lastFiredDate\":null},\"mentalCheck\":{\"mood\":5,\"stress\":5,\"anxiety\":4,\"sleepQuality\":\"good\",\"support\":\"no\"},\"mentalScores\":[]}', '2026-08-01 05:55:56');

-- --------------------------------------------------------

--
-- Table structure for table `profiles`
--

CREATE TABLE `profiles` (
  `user_id` bigint(20) UNSIGNED NOT NULL,
  `full_name` varchar(160) NOT NULL,
  `age` int(11) NOT NULL DEFAULT 0,
  `gender` varchar(20) NOT NULL DEFAULT 'male',
  `village` varchar(160) NOT NULL DEFAULT '',
  `city` varchar(160) NOT NULL DEFAULT '',
  `phone` varchar(40) NOT NULL DEFAULT '',
  `updated_at` timestamp NOT NULL DEFAULT current_timestamp() ON UPDATE current_timestamp()
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

--
-- Dumping data for table `profiles`
--

INSERT INTO `profiles` (`user_id`, `full_name`, `age`, `gender`, `village`, `city`, `phone`, `updated_at`) VALUES
(2, 'KFC grandpa', 100, 'male', 'Kampung Bunuk', 'Kuching', '+60199894341', '2026-08-01 05:50:20');

-- --------------------------------------------------------

--
-- Table structure for table `sessions`
--

CREATE TABLE `sessions` (
  `id` bigint(20) UNSIGNED NOT NULL,
  `user_id` bigint(20) UNSIGNED NOT NULL,
  `token_hash` char(64) NOT NULL,
  `expires_at` datetime NOT NULL,
  `created_at` timestamp NOT NULL DEFAULT current_timestamp()
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

--
-- Dumping data for table `sessions`
--

INSERT INTO `sessions` (`id`, `user_id`, `token_hash`, `expires_at`, `created_at`) VALUES
(3, 2, 'dd9336262225c5673c9cf0a94921276f7abbf4b9bde487e4d495433360695d43', '2026-08-08 05:46:02', '2026-08-01 05:46:02');

-- --------------------------------------------------------

--
-- Table structure for table `users`
--

CREATE TABLE `users` (
  `id` bigint(20) UNSIGNED NOT NULL,
  `email` varchar(255) NOT NULL,
  `password_hash` varchar(255) NOT NULL,
  `created_at` timestamp NOT NULL DEFAULT current_timestamp(),
  `updated_at` timestamp NOT NULL DEFAULT current_timestamp() ON UPDATE current_timestamp()
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

--
-- Dumping data for table `users`
--

INSERT INTO `users` (`id`, `email`, `password_hash`, `created_at`, `updated_at`) VALUES
(2, 'admin@gmail.com', 'pbkdf2_sha256$120000$gXKTAlWuaQJpB9ZgjeZnoQ$UPNCUfgJN7mgnn1CHhc+WhKUCOV5i1i5/wxU/e86zRk', '2026-08-01 05:46:02', '2026-08-01 05:46:02');

--
-- Indexes for dumped tables
--

--
-- Indexes for table `app_states`
--
ALTER TABLE `app_states`
  ADD PRIMARY KEY (`user_id`);

--
-- Indexes for table `profiles`
--
ALTER TABLE `profiles`
  ADD PRIMARY KEY (`user_id`);

--
-- Indexes for table `sessions`
--
ALTER TABLE `sessions`
  ADD PRIMARY KEY (`id`),
  ADD UNIQUE KEY `token_hash` (`token_hash`),
  ADD KEY `idx_sessions_user` (`user_id`),
  ADD KEY `idx_sessions_expires` (`expires_at`);

--
-- Indexes for table `users`
--
ALTER TABLE `users`
  ADD PRIMARY KEY (`id`),
  ADD UNIQUE KEY `email` (`email`);

--
-- AUTO_INCREMENT for dumped tables
--

--
-- AUTO_INCREMENT for table `sessions`
--
ALTER TABLE `sessions`
  MODIFY `id` bigint(20) UNSIGNED NOT NULL AUTO_INCREMENT, AUTO_INCREMENT=4;

--
-- AUTO_INCREMENT for table `users`
--
ALTER TABLE `users`
  MODIFY `id` bigint(20) UNSIGNED NOT NULL AUTO_INCREMENT, AUTO_INCREMENT=3;

--
-- Constraints for dumped tables
--

--
-- Constraints for table `app_states`
--
ALTER TABLE `app_states`
  ADD CONSTRAINT `fk_app_states_user` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE CASCADE;

--
-- Constraints for table `profiles`
--
ALTER TABLE `profiles`
  ADD CONSTRAINT `fk_profiles_user` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE CASCADE;

--
-- Constraints for table `sessions`
--
ALTER TABLE `sessions`
  ADD CONSTRAINT `fk_sessions_user` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE CASCADE;
COMMIT;

/*!40101 SET CHARACTER_SET_CLIENT=@OLD_CHARACTER_SET_CLIENT */;
/*!40101 SET CHARACTER_SET_RESULTS=@OLD_CHARACTER_SET_RESULTS */;
/*!40101 SET COLLATION_CONNECTION=@OLD_COLLATION_CONNECTION */;
