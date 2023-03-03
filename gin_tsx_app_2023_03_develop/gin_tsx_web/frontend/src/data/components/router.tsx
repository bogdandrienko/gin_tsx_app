// TODO download modules ///////////////////////////////////////////////////////////////////////////////////////////////

import React from "react";
import { BrowserRouter, Routes, Route } from "react-router-dom";
import axios from "axios";

// TODO custom modules /////////////////////////////////////////////////////////////////////////////////////////////////

import HomePage from "../pages/HomePage";
import DumptrucksScan from "../screens/DumptrucksScan";
import LoginPage from "../pages/LoginPage";
import LogoutPage from "../pages/LogoutPage";
import RegisterPage from "../pages/RegisterPage";
import TaskListPage from "../pages/TaskListPage";
import TaskPage from "../pages/TaskPage";
import TaskCreatePage from "../pages/TaskCreatePage";
import TaskChangePage from "../pages/TaskChangePage";

import "../css/bootstrap/bootstrap.min.css";
import "../css/font_awesome/css/all.min.css";
import "../css/my.css";

import * as constants from "./constants";

// TODO settings ///////////////////////////////////////////////////////////////////////////////////////////////////////

axios.defaults.baseURL = constants.SERVER_HOST_AND_PORT_CONSTANT;
axios.defaults.headers.common = {
  ...axios.defaults.headers.common,
  // "Access-Control-Allow-Origin": "http://127.0.0.1:3000",
  "Access-Control-Allow-Origin": "*",
};

// TODO export /////////////////////////////////////////////////////////////////////////////////////////////////////////

export default function Page() {
  // TODO return ///////////////////////////////////////////////////////////////////////////////////////////////////////
  return (
    <BrowserRouter>
      <Routes>
        <Route path="/" element={<DumptrucksScan />}></Route>
        {/*<Route path="/" element={<HomePage />}></Route>*/}
        {/*<Route path="/login" element={<LoginPage />}></Route>*/}
        {/*<Route path="/logout" element={<LogoutPage />}></Route>*/}
        {/*<Route path="/register" element={<RegisterPage />}></Route>*/}
        {/*<Route path="/tasks" element={<TaskListPage />}></Route>*/}
        {/*<Route path="/tasks/:id" element={<TaskPage />}></Route>*/}
        {/*<Route path="/tasks/create" element={<TaskCreatePage />}></Route>*/}
        {/*<Route path="/tasks/update/:id" element={<TaskChangePage />}></Route>*/}
      </Routes>
    </BrowserRouter>
  );
}
