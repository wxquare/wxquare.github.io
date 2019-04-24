---
title: Python Logging 日志记录
categories: 
- Python
---

Python的logging模块非常灵活，提供许多的配置项，方便定义所需的日志管理器，我认为日志至少应做到以下几点：
1. 灵活日志等级
2. 日志的格式，包含时间、文件、文件行数、日志信息等。（logging.Formater)
3. 日志以什么方式写到哪里？(logging.handlers)
结合官方文档可以轻松定义所需要的日志记录器：
[https://docs.python.org/zh-cn/3/howto/logging.html](https://docs.python.org/zh-cn/3/howto/logging.html)

```
	import logging
	import logging.handlers


	def get_logger(name="log",level=logging.DEBUG,console_switch=False):
	    # create logger
	    logger = logging.getLogger(name)
	    logger.setLevel(level)

	    # create formatter
	    log_format=logging.Formatter("%(asctime)s|%(levelname)s|%(filename)s[%(lineno)d]|%(message)s")

	    # create TimedRotatingFileHandler
	    rf_handler = logging.handlers.TimedRotatingFileHandler('all.log', when='midnight', interval=1, backupCount=7)
	    rf_handler.setFormatter(log_format)
	    rf_handler.setLevel(logging.DEBUG)

	    # create FileHandler
	    f_handler = logging.FileHandler('error.log')
	    f_handler.setLevel(logging.ERROR)
	    f_handler.setFormatter(log_format)

	    # create console handler with a higher log level
	    if console_switch:
	        ch = logging.StreamHandler()
	        ch.setLevel(logging.DEBUG)
	        ch.setFormatter(log_format)
	        logger.addHandler(ch)

	    # add handles to logger
	    logger.addHandler(rf_handler)
	    logger.addHandler(f_handler)
	    return logger

	if __name__ == '__main__':
	    
	    logger = get_logger(console_switch=True)
	    logger.debug("debug message")
	    logger.error("error message")

```