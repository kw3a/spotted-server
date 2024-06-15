-- MySQL dump 10.13  Distrib 8.3.0, for Linux (x86_64)
--
-- Host: localhost    Database: spotted-db
-- ------------------------------------------------------
-- Server version	8.3.0

/*!40101 SET @OLD_CHARACTER_SET_CLIENT=@@CHARACTER_SET_CLIENT */;
/*!40101 SET @OLD_CHARACTER_SET_RESULTS=@@CHARACTER_SET_RESULTS */;
/*!40101 SET @OLD_COLLATION_CONNECTION=@@COLLATION_CONNECTION */;
/*!50503 SET NAMES utf8mb4 */;
/*!40103 SET @OLD_TIME_ZONE=@@TIME_ZONE */;
/*!40103 SET TIME_ZONE='+00:00' */;
/*!40014 SET @OLD_UNIQUE_CHECKS=@@UNIQUE_CHECKS, UNIQUE_CHECKS=0 */;
/*!40014 SET @OLD_FOREIGN_KEY_CHECKS=@@FOREIGN_KEY_CHECKS, FOREIGN_KEY_CHECKS=0 */;
/*!40101 SET @OLD_SQL_MODE=@@SQL_MODE, SQL_MODE='NO_AUTO_VALUE_ON_ZERO' */;
/*!40111 SET @OLD_SQL_NOTES=@@SQL_NOTES, SQL_NOTES=0 */;

--
-- Table structure for table `best_try`
--

DROP TABLE IF EXISTS `best_try`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `best_try` (
  `id` char(36) NOT NULL,
  `accepted_test_cases` tinyint unsigned NOT NULL DEFAULT '0',
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `participation_id` char(36) NOT NULL,
  `problem_id` char(36) NOT NULL,
  `submission_id` char(36) DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `participation_id` (`participation_id`,`problem_id`),
  KEY `problem_id` (`problem_id`),
  KEY `submission_id` (`submission_id`),
  CONSTRAINT `best_try_ibfk_1` FOREIGN KEY (`participation_id`) REFERENCES `participation` (`id`) ON DELETE CASCADE,
  CONSTRAINT `best_try_ibfk_2` FOREIGN KEY (`problem_id`) REFERENCES `problem` (`id`) ON DELETE CASCADE,
  CONSTRAINT `best_try_ibfk_3` FOREIGN KEY (`submission_id`) REFERENCES `submission` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `example`
--

DROP TABLE IF EXISTS `example`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `example` (
  `id` char(36) NOT NULL,
  `input` varchar(255) NOT NULL,
  `output` varchar(255) NOT NULL,
  `problem_id` char(36) NOT NULL,
  PRIMARY KEY (`id`),
  KEY `problem_id` (`problem_id`),
  CONSTRAINT `example_ibfk_1` FOREIGN KEY (`problem_id`) REFERENCES `problem` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `goose_db_version`
--

DROP TABLE IF EXISTS `goose_db_version`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `goose_db_version` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `version_id` bigint NOT NULL,
  `is_applied` tinyint(1) NOT NULL,
  `tstamp` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `id` (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=199 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `jwt`
--

DROP TABLE IF EXISTS `jwt`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `jwt` (
  `refresh_token` varchar(256) NOT NULL,
  `created_at` timestamp NOT NULL,
  PRIMARY KEY (`refresh_token`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `language`
--

DROP TABLE IF EXISTS `language`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `language` (
  `id` int NOT NULL,
  `name` varchar(255) NOT NULL,
  `version` int NOT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `language_problem`
--

DROP TABLE IF EXISTS `language_problem`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `language_problem` (
  `id` char(36) NOT NULL,
  `language_id` int NOT NULL,
  `problem_id` char(36) NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `problem_id` (`problem_id`,`language_id`),
  KEY `language_id` (`language_id`),
  CONSTRAINT `language_problem_ibfk_1` FOREIGN KEY (`language_id`) REFERENCES `language` (`id`) ON DELETE CASCADE,
  CONSTRAINT `language_problem_ibfk_2` FOREIGN KEY (`problem_id`) REFERENCES `problem` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `participation`
--

DROP TABLE IF EXISTS `participation`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `participation` (
  `id` char(36) NOT NULL,
  `date` timestamp NOT NULL,
  `user_id` char(36) NOT NULL,
  `quiz_id` char(36) NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `user_id` (`user_id`,`quiz_id`),
  KEY `quiz_id` (`quiz_id`),
  CONSTRAINT `participation_ibfk_1` FOREIGN KEY (`user_id`) REFERENCES `user` (`id`) ON DELETE CASCADE,
  CONSTRAINT `participation_ibfk_2` FOREIGN KEY (`quiz_id`) REFERENCES `quiz` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
/*!40101 SET character_set_client = @saved_cs_client */;
/*!50003 SET @saved_cs_client      = @@character_set_client */ ;
/*!50003 SET @saved_cs_results     = @@character_set_results */ ;
/*!50003 SET @saved_col_connection = @@collation_connection */ ;
/*!50003 SET character_set_client  = utf8mb4 */ ;
/*!50003 SET character_set_results = utf8mb4 */ ;
/*!50003 SET collation_connection  = utf8mb4_general_ci */ ;
/*!50003 SET @saved_sql_mode       = @@sql_mode */ ;
/*!50003 SET sql_mode              = 'ONLY_FULL_GROUP_BY,STRICT_TRANS_TABLES,NO_ZERO_IN_DATE,NO_ZERO_DATE,ERROR_FOR_DIVISION_BY_ZERO,NO_ENGINE_SUBSTITUTION' */ ;
DELIMITER ;;
/*!50003 CREATE*/ /*!50017 DEFINER=`root`@`%`*/ /*!50003 TRIGGER `insert_best_try_participation` AFTER INSERT ON `participation` FOR EACH ROW BEGIN
    DECLARE problem_id INT;
    DECLARE cursor_finished BOOLEAN DEFAULT FALSE;

    DECLARE cur_problems CURSOR FOR 
        SELECT id
        FROM problem
        WHERE quiz_id = (
            SELECT quiz_id
            FROM participation
            WHERE id = NEW.id
        );

    DECLARE CONTINUE HANDLER FOR NOT FOUND 
        SET cursor_finished = TRUE;

    OPEN cur_problems;

    loop_problems: LOOP
        FETCH cur_problems INTO problem_id;
        IF cursor_finished THEN
            LEAVE loop_problems;
        END IF; -- IF END

        INSERT INTO best_try (participation_id, problem_id)
        VALUES (NEW.id, problem_id);
    END LOOP loop_problems; -- LOOP END

    CLOSE cur_problems; -- CURSOR END
END */;;
DELIMITER ;
/*!50003 SET sql_mode              = @saved_sql_mode */ ;
/*!50003 SET character_set_client  = @saved_cs_client */ ;
/*!50003 SET character_set_results = @saved_cs_results */ ;
/*!50003 SET collation_connection  = @saved_col_connection */ ;

--
-- Table structure for table `problem`
--

DROP TABLE IF EXISTS `problem`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `problem` (
  `id` char(36) NOT NULL,
  `description` text NOT NULL,
  `title` varchar(255) NOT NULL,
  `memory_limit` int NOT NULL,
  `time_limit` double NOT NULL,
  `quiz_id` char(36) NOT NULL,
  PRIMARY KEY (`id`),
  KEY `quiz_id` (`quiz_id`),
  CONSTRAINT `problem_ibfk_1` FOREIGN KEY (`quiz_id`) REFERENCES `quiz` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `quiz`
--

DROP TABLE IF EXISTS `quiz`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `quiz` (
  `id` char(36) NOT NULL,
  `title` varchar(255) NOT NULL,
  `description` varchar(255) NOT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `submission`
--

DROP TABLE IF EXISTS `submission`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `submission` (
  `id` char(36) NOT NULL,
  `src` text NOT NULL,
  `time` timestamp NOT NULL,
  `accepted_test_cases` tinyint unsigned NOT NULL DEFAULT '0',
  `problem_id` char(36) NOT NULL,
  `participation_id` char(36) NOT NULL,
  `language_id` int NOT NULL,
  PRIMARY KEY (`id`),
  KEY `problem_id` (`problem_id`),
  KEY `participation_id` (`participation_id`),
  KEY `language_id` (`language_id`),
  CONSTRAINT `submission_ibfk_1` FOREIGN KEY (`problem_id`) REFERENCES `problem` (`id`) ON DELETE CASCADE,
  CONSTRAINT `submission_ibfk_2` FOREIGN KEY (`participation_id`) REFERENCES `participation` (`id`) ON DELETE CASCADE,
  CONSTRAINT `submission_ibfk_3` FOREIGN KEY (`language_id`) REFERENCES `language` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
/*!40101 SET character_set_client = @saved_cs_client */;
/*!50003 SET @saved_cs_client      = @@character_set_client */ ;
/*!50003 SET @saved_cs_results     = @@character_set_results */ ;
/*!50003 SET @saved_col_connection = @@collation_connection */ ;
/*!50003 SET character_set_client  = utf8mb4 */ ;
/*!50003 SET character_set_results = utf8mb4 */ ;
/*!50003 SET collation_connection  = utf8mb4_general_ci */ ;
/*!50003 SET @saved_sql_mode       = @@sql_mode */ ;
/*!50003 SET sql_mode              = 'ONLY_FULL_GROUP_BY,STRICT_TRANS_TABLES,NO_ZERO_IN_DATE,NO_ZERO_DATE,ERROR_FOR_DIVISION_BY_ZERO,NO_ENGINE_SUBSTITUTION' */ ;
DELIMITER ;;
/*!50003 CREATE*/ /*!50017 DEFINER=`root`@`%`*/ /*!50003 TRIGGER `insert_test_case_result_submission` AFTER INSERT ON `submission` FOR EACH ROW BEGIN
    DECLARE test_case_id INT;
    DECLARE cursor_finished BOOLEAN DEFAULT FALSE;

    DECLARE cur_test_cases CURSOR FOR 
        SELECT id
        FROM test_case
        WHERE problem_id = NEW.problem_id;

    DECLARE CONTINUE HANDLER FOR NOT FOUND 
        SET cursor_finished = TRUE;

    OPEN cur_test_cases;

    loop_test_cases: LOOP
        FETCH cur_test_cases INTO test_case_id;
        IF cursor_finished THEN
            LEAVE loop_test_cases;
        END IF; -- IF END

        INSERT INTO test_case_result (problem_id, submission_id, test_case_id)
        VALUES (NEW.problem_id, NEW.id, test_case_id);
    END LOOP loop_test_cases; -- LOOP END

    CLOSE cur_test_cases; -- CURSOR END
END */;;
DELIMITER ;
/*!50003 SET sql_mode              = @saved_sql_mode */ ;
/*!50003 SET character_set_client  = @saved_cs_client */ ;
/*!50003 SET character_set_results = @saved_cs_results */ ;
/*!50003 SET collation_connection  = @saved_col_connection */ ;
/*!50003 SET @saved_cs_client      = @@character_set_client */ ;
/*!50003 SET @saved_cs_results     = @@character_set_results */ ;
/*!50003 SET @saved_col_connection = @@collation_connection */ ;
/*!50003 SET character_set_client  = utf8mb4 */ ;
/*!50003 SET character_set_results = utf8mb4 */ ;
/*!50003 SET collation_connection  = utf8mb4_general_ci */ ;
/*!50003 SET @saved_sql_mode       = @@sql_mode */ ;
/*!50003 SET sql_mode              = 'ONLY_FULL_GROUP_BY,STRICT_TRANS_TABLES,NO_ZERO_IN_DATE,NO_ZERO_DATE,ERROR_FOR_DIVISION_BY_ZERO,NO_ENGINE_SUBSTITUTION' */ ;
DELIMITER ;;
/*!50003 CREATE*/ /*!50017 DEFINER=`root`@`%`*/ /*!50003 TRIGGER `update_best_try_submission` AFTER UPDATE ON `submission` FOR EACH ROW BEGIN
    IF NEW.accepted_test_cases > best_try.accepted_test_cases  THEN
        UPDATE best_try
        SET accepted_test_cases = NEW.accepted_test_cases 
        AND submission_id = NEW.id
        WHERE participation_id = NEW.participation_id 
        AND problem_id = NEW.problem_id;
    END IF; -- IF END
END */;;
DELIMITER ;
/*!50003 SET sql_mode              = @saved_sql_mode */ ;
/*!50003 SET character_set_client  = @saved_cs_client */ ;
/*!50003 SET character_set_results = @saved_cs_results */ ;
/*!50003 SET collation_connection  = @saved_col_connection */ ;

--
-- Table structure for table `test_case`
--

DROP TABLE IF EXISTS `test_case`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `test_case` (
  `id` char(36) NOT NULL,
  `input` varchar(255) NOT NULL,
  `output` varchar(255) NOT NULL,
  `problem_id` char(36) NOT NULL,
  PRIMARY KEY (`id`),
  KEY `problem_id` (`problem_id`),
  CONSTRAINT `test_case_ibfk_1` FOREIGN KEY (`problem_id`) REFERENCES `problem` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `test_case_result`
--

DROP TABLE IF EXISTS `test_case_result`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `test_case_result` (
  `id` char(36) NOT NULL,
  `status` varchar(64) NOT NULL,
  `time` decimal(6,3) NOT NULL,
  `memory` int NOT NULL,
  `test_case_id` char(36) NOT NULL,
  `submission_id` char(36) NOT NULL,
  PRIMARY KEY (`id`),
  KEY `test_case_id` (`test_case_id`),
  KEY `submission_id` (`submission_id`),
  CONSTRAINT `test_case_result_ibfk_1` FOREIGN KEY (`test_case_id`) REFERENCES `test_case` (`id`) ON DELETE CASCADE,
  CONSTRAINT `test_case_result_ibfk_2` FOREIGN KEY (`submission_id`) REFERENCES `submission` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
/*!40101 SET character_set_client = @saved_cs_client */;
/*!50003 SET @saved_cs_client      = @@character_set_client */ ;
/*!50003 SET @saved_cs_results     = @@character_set_results */ ;
/*!50003 SET @saved_col_connection = @@collation_connection */ ;
/*!50003 SET character_set_client  = utf8mb4 */ ;
/*!50003 SET character_set_results = utf8mb4 */ ;
/*!50003 SET collation_connection  = utf8mb4_general_ci */ ;
/*!50003 SET @saved_sql_mode       = @@sql_mode */ ;
/*!50003 SET sql_mode              = 'ONLY_FULL_GROUP_BY,STRICT_TRANS_TABLES,NO_ZERO_IN_DATE,NO_ZERO_DATE,ERROR_FOR_DIVISION_BY_ZERO,NO_ENGINE_SUBSTITUTION' */ ;
DELIMITER ;;
/*!50003 CREATE*/ /*!50017 DEFINER=`root`@`%`*/ /*!50003 TRIGGER `update_submission_test_cases` AFTER UPDATE ON `test_case_result` FOR EACH ROW BEGIN
    IF NEW.status = 'Accepted' THEN
        UPDATE submission
        SET accepted_test_cases = accepted_test_cases + 1
        WHERE id = NEW.submission_id;
    END IF; -- IF END
END */;;
DELIMITER ;
/*!50003 SET sql_mode              = @saved_sql_mode */ ;
/*!50003 SET character_set_client  = @saved_cs_client */ ;
/*!50003 SET character_set_results = @saved_cs_results */ ;
/*!50003 SET collation_connection  = @saved_col_connection */ ;

--
-- Table structure for table `user`
--

DROP TABLE IF EXISTS `user`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `user` (
  `id` char(36) NOT NULL,
  `name` varchar(255) NOT NULL,
  `email` varchar(255) NOT NULL,
  `password` varchar(255) NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `email` (`email`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
/*!40101 SET character_set_client = @saved_cs_client */;
/*!40103 SET TIME_ZONE=@OLD_TIME_ZONE */;

/*!40101 SET SQL_MODE=@OLD_SQL_MODE */;
/*!40014 SET FOREIGN_KEY_CHECKS=@OLD_FOREIGN_KEY_CHECKS */;
/*!40014 SET UNIQUE_CHECKS=@OLD_UNIQUE_CHECKS */;
/*!40101 SET CHARACTER_SET_CLIENT=@OLD_CHARACTER_SET_CLIENT */;
/*!40101 SET CHARACTER_SET_RESULTS=@OLD_CHARACTER_SET_RESULTS */;
/*!40101 SET COLLATION_CONNECTION=@OLD_COLLATION_CONNECTION */;
/*!40111 SET SQL_NOTES=@OLD_SQL_NOTES */;

-- Dump completed on 2024-05-16 20:39:33
