---
title: 电商支付系统深度解析
date: 2026-04-09
categories:
  - system-design
tags:
  - payment
  - distributed-transaction
  - idempotency
  - state-machine
  - saga
  - tcc
  - reconciliation
---

## 引言

支付系统是电商平台的资金流枢纽，连接用户、平台、商家、第三方支付等多方角色。本文从系统设计面试的角度，深入解析支付系统的核心流程、状态机设计、分布式事务等高频考点。

**适合读者**：准备系统设计面试的候选人

**阅读时长**：30-40 分钟

**核心内容**：
- 支付系统整体架构
- 支付和退款流程
- 状态机设计
- 分布式事务（Saga/TCC）
- 幂等性设计
- 一致性保证
