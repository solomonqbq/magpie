-- phpMyAdmin SQL Dump
-- version 4.0.10deb1
-- http://www.phpmyadmin.net
--
-- 主机: 192.168.15.64:3306
-- 生成日期: 2015-04-02 16:10:19
-- 服务器版本: 5.6.17-log
-- PHP 版本: 5.5.9-1ubuntu4.6

SET SQL_MODE = "NO_AUTO_VALUE_ON_ZERO";
SET time_zone = "+00:00";


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
(1, 'test_group', '2015-04-02 06:18:29', '2015-04-02 07:41:02');

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
  KEY `worker_id` (`worker_id`),
  KEY `group` (`group`)
) ENGINE=InnoDB  DEFAULT CHARSET=utf8 AUTO_INCREMENT=7 ;

--
-- 转存表中的数据 `mp_task`
--

INSERT INTO `mp_task` (`id`, `name`, `context`, `group`, `worker_id`, `retry`, `run_type`, `interval`, `exception`, `created_time`, `updated_tIme`, `status`) VALUES
(1, 'test0', 'a=1,b=2', 'test_group', 114, 0, 0, 0, 'mock error', '0000-00-00 00:00:00', '2015-04-02 07:50:58', 0),
(2, 'test1', 'a=1,b=2', 'test_group', 114, 0, 0, 0, 'mock error', '0000-00-00 00:00:00', '2015-04-02 07:50:58', 0),
(3, 'test2', 'a=1,b=2', 'test_group', 113, 0, 0, 0, 'mock error', '0000-00-00 00:00:00', '2015-04-02 07:50:58', 0),
(4, 'test3', 'a=51,b=25', 'test_group', 113, 0, 0, 0, 'mock error', '0000-00-00 00:00:00', '2015-04-02 07:50:58', 0),
(5, 'test4', 'a=21,b=277', 'test_group', 113, 0, 0, 0, '', '0000-00-00 00:00:00', '2015-04-02 07:50:58', 0),
(6, 'test5', 'a=1,b=http://baidu.com', 'test_group', 113, 0, 0, 0, 'mock error', '0000-00-00 00:00:00', '2015-04-02 07:50:58', 0);

-- --------------------------------------------------------

--
-- 表的结构 `mp_worker`
--

CREATE TABLE IF NOT EXISTS `mp_worker` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT COMMENT '自增主键',
  `name` varchar(32) NOT NULL DEFAULT '',
  `group` varchar(32) NOT NULL COMMENT '组名',
  `created_time` timestamp NOT NULL DEFAULT '0000-00-00 00:00:00' COMMENT '创建时间',
  `updated_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '最后更新时间',
  `time_out` timestamp NOT NULL DEFAULT '0000-00-00 00:00:00' COMMENT '到期有效时',
  PRIMARY KEY (`id`),
  KEY `group` (`group`)
) ENGINE=InnoDB  DEFAULT CHARSET=utf8 AUTO_INCREMENT=115 ;

--
-- 转存表中的数据 `mp_worker`
--

INSERT INTO `mp_worker` (`id`, `name`, `group`, `created_time`, `updated_time`, `time_out`) VALUES
(114, '10.12.121.72', 'test_group', '2015-04-02 08:07:28', '2015-04-02 07:49:21', '2015-04-02 07:49:31');

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
) ENGINE=InnoDB  DEFAULT CHARSET=utf8 AUTO_INCREMENT=64 ;

--
-- 转存表中的数据 `mp_worker_group`
--

INSERT INTO `mp_worker_group` (`id`, `group`, `worker_id`, `time_out`, `created_time`, `updated_time`) VALUES
(63, 'test_group', 114, '2015-04-02 07:49:31', '0000-00-00 00:00:00', '2015-04-02 07:49:21');

/*!40101 SET CHARACTER_SET_CLIENT=@OLD_CHARACTER_SET_CLIENT */;
/*!40101 SET CHARACTER_SET_RESULTS=@OLD_CHARACTER_SET_RESULTS */;
/*!40101 SET COLLATION_CONNECTION=@OLD_COLLATION_CONNECTION */;