// TODO download modules ///////////////////////////////////////////////////////////////////////////////////////////////

import React, { useEffect } from "react";
import axios from "axios";

// TODO custom modules /////////////////////////////////////////////////////////////////////////////////////////////////

import * as base from "../components/ui/base";
import * as hooks from "../components/hooks";
import { Link } from "react-router-dom";

// TODO export /////////////////////////////////////////////////////////////////////////////////////////////////////////

export default function Page(): JSX.Element {
  // TODO return ///////////////////////////////////////////////////////////////////////////////////////////////////////
  return (
    <base.Base1
      title={"Домашняя страница"}
      description={"тут расположены ссылки на основные модули"}
    >
      <div className={"input-group lead my-5"}>
        <Link
          to={"/dumptrucks"}
          className={"btn btn-lg btn-outline-primary fw-bold lead display-6"}
        >
          Анализ автосамосвалов
        </Link>
        <Link
          to={"/shovels"}
          className={"btn btn-lg btn-outline-success fw-bold lead display-6"}
        >
          Анализ экскаваторов
        </Link>
        <Link
          to={"/predictive"}
          className={"btn btn-lg btn-outline-warning fw-bold lead display-6"}
        >
          Предиктивная аналитика
        </Link>
      </div>
    </base.Base1>
  );
}
