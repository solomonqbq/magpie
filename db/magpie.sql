-- phpMyAdmin SQL Dump
-- version 4.0.10deb1
-- http://www.phpmyadmin.net
--
-- 主机: 192.168.15.64:3306
-- 生成日期: 2015-04-01 18:51:59
-- 服务器版本: 5.6.17-log
-- PHP 版本: 5.5.9-1ubuntu4.6

SET SQL_MODE = "NO_AUTO_VALUE_ON_ZERO";
SET time_zone = "+00:00";

d
/*!40101 SET @OLD_CHARACTER_SET_CLIENT=@@CHARACTER_SET_CLIENT */;
/*!40101 SET @OLD_CHARACTER_SET_RESULTS=@@CHARACTER_SET_RESULTS */;
/*!40101 SET @OLD_COLLATION_CONNECTION=@@COLLATION_CONNECTION */;
/*!40101 SET NAMES utf8 */;

--
-- 数据库: `magpie`
--

-- --------------------------------------------------------

--
-- 表的结构 `mp_group`
--

CREATE TABLE IF NOT EXISTS `mp_group` (
  `id` int(11) NOT NULL AUTO_INCREMENT COMMENT '自增主键',
  `group` varchar(32) NOT NULL COMMENT '组名',
  `created_time` timestamp NOT NULL DEFAULT '0000-00-00 00:00:00' COMMENT '创建时间',
  `updated_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `group` (`group`)
) ENGINE=InnoDB  DEFAULT CHARSET=utf8 AUTO_INCREMENT=2 ;

--
-- 转存表中的数据 `mp_group`
--

INSERT INTO `mp_group` (`id`, `group`, `created_time`, `updated_time`) VALUES
(1, 'load_balance', '2015-04-01 09:32:21', '2015-04-01 09:32:21');

-- --------------------------------------------------------

--
-- 表的结构 `mp_task`
--

CREATE TABLE IF NOT EXISTS `mp_task` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT '自增ID',
  `name` varchar(32) DEFAULT '“”' COMMENT '任务名',
  `context` varchar(256) DEFAULT '""' COMMENT '任务上下文参数，逗号分割，key=value的形式,如a=1,b=http://baidu.com.',
  `group` varchar(32) NOT NULL COMMENT '组名',
  `worker_id` bigint(20) NOT NULL DEFAULT '-1' COMMENT '所有者的workerID',
  `retry` int(11) NOT NULL DEFAULT '0' COMMENT '重试次数',
  `run_type` int(11) NOT NULL COMMENT '任务类型 0:周期任务 1:一次性任务',
  `interval` int(11) NOT NULL COMMENT '运行间隔，只有run_type是周期任务时才有效',
  `exception` varchar(512) DEFAULT NULL COMMENT '最后一次错误',
  `created_time` timestamp NOT NULL DEFAULT '0000-00-00 00:00:00' COMMENT '创建时间',
  `updated_tIme` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '最后更新时间',
  `status` tinyint(4) NOT NULL DEFAULT '0' COMMENT '0:新任务 1:已分配 2:运行中 3::失败 4:成功 5:错误',
  PRIMARY KEY (`id`),
  KEY `status` (`status`),
  KEY `group` (`group`)
) ENGINE=InnoDB  DEFAULT CHARSET=utf8 AUTO_INCREMENT=7 ;

--
-- 转存表中的数据 `mp_task`
--

INSERT INTO `mp_task` (`id`, `name`, `context`, `group`, `worker_id`, `retry`, `run_type`, `interval`, `exception`, `created_time`, `updated_tIme`, `status`) VALUES
(1, 'deliver_conf', '""', 'load_balance', 65, 0, 0, 0, NULL, '0000-00-00 00:00:00', '2015-04-01 09:40:13', 2),
(2, 'test', '""', 'load_balance', 65, 0, 0, 0, NULL, '0000-00-00 00:00:00', '2015-04-01 09:33:48', 2),
(3, 'test2', '""', 'load_balance', 65, 0, 0, 0, NULL, '0000-00-00 00:00:00', '2015-04-01 09:40:08', 2),
(4, 'test3', '""', 'load_balance', 66, 0, 0, 0, NULL, '0000-00-00 00:00:00', '2015-04-01 09:33:46', 2),
(5, 'test4', '""', 'load_balance', 66, 0, 0, 0, NULL, '0000-00-00 00:00:00', '2015-04-01 09:40:17', 2),
(6, 'test5', '""', 'load_balance', 66, 0, 0, 0, NULL, '0000-00-00 00:00:00', '2015-04-01 09:40:12', 2);

-- --------------------------------------------------------

--
-- 表的结构 `mp_worker`
--

CREATE TABLE IF NOT EXISTS `mp_worker` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT COMMENT '自增主键',
  `name` varchar(32) NOT NULL DEFAULT '',
  `created_time` timestamp NOT NULL DEFAULT '0000-00-00 00:00:00' COMMENT '创建时间',
  `updated_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '最后更新时间',
  `time_out` timestamp NOT NULL DEFAULT '0000-00-00 00:00:00' COMMENT '到期有效时',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB  DEFAULT CHARSET=utf8 AUTO_INCREMENT=67 ;

--
-- 转存表中的数据 `mp_worker`
--

INSERT INTO `mp_worker` (`id`, `name`, `created_time`, `updated_time`, `time_out`) VALUES
(65, '10.12.121.72', '2015-04-01 09:51:11', '2015-04-01 09:42:13', '2015-04-01 09:42:23'),
(66, '10.12.121.72', '2015-04-01 09:51:15', '2015-04-01 09:42:17', '2015-04-01 09:42:27');

-- --------------------------------------------------------

--
-- 表的结构 `mp_worker_group`
--

CREATE TABLE IF NOT EXISTS `mp_worker_group` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT COMMENT '自增主键',
  `group` varchar(32) NOT NULL COMMENT '组名',
  `worker_id` bigint(20) NOT NULL COMMENT '当前组的leader看板id',
  `time_out` timestamp NOT NULL DEFAULT '0000-00-00 00:00:00' COMMENT '到期有效时间',
  `created_time` timestamp NOT NULL DEFAULT '0000-00-00 00:00:00' COMMENT '创建时间',
  `updated_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '最后更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `group` (`group`)
) ENGINE=InnoDB  DEFAULT CHARSET=utf8 AUTO_INCREMENT=52 ;

--
-- 转存表中的数据 `mp_worker_group`
--

INSERT INTO `mp_worker_group` (`id`, `group`, `worker_id`, `time_out`, `created_time`, `updated_time`) VALUES
(1, 'load_balance', 66, '2015-04-01 09:42:27', '2015-04-01 04:59:38', '2015-04-01 09:42:17');

/*!40101 SET CHARACTER_SET_CLIENT=@OLD_CHARACTER_SET_CLIENT */;
/*!40101 SET CHARACTER_SET_RESULTS=@OLD_CHARACTER_SET_RESULTS */;
/*!40101 SET COLLATION_CONNECTION=@OLD_COLLATION_CONNECTION */;
