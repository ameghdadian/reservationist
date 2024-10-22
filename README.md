<div align="center">
    <h1>Reservationist</h1>
    <a href="https://github.com/ameghdadian/reservationist/actions">
        <img src="https://img.shields.io/github/actions/workflow/status/ameghdadian/reservationist/github-actions.yaml?branch=main&label=CI&logo=github&style=flat-square" height="20" alt="GitHub Workflow Status">
  </a>
    <a href="https://img.shields.io/coverallsCoverage/github/ameghdadian/reservationist">
        <img src="https://img.shields.io/coverallsCoverage/github/ameghdadian/reservationist" height="20" alt="GitHub Workflow Status">
  </a>

  <h3>A Restful API written in Go, confirming with DDD principles which you can use to book appointments</h3>
</div>

`Reservationist` is an application that makes connection between businesses and customers much easier and faster. By using `Reservationist`, customers can easily browse through working agenda of a business and make appointments. It also supports sending early notification to customers reminding them of their appointments.

## Technical Document

Please find technical documentations/diagrams and design decisions described in the [docs](./docs/) directory.

## System Overview
![System in a glimpse](./docs/arch/System-Overview.svg)

## Run Locally

To run this application locally, first make sure you have `Docker` installed locallay and then follow these steps:
1. make dev-docker
2. make dev-up
3. make dev-update-apply

