// TODO download modules ///////////////////////////////////////////////////////////////////////////////////////////////

import React, { useEffect } from "react";
import axios from "axios";

// TODO custom modules /////////////////////////////////////////////////////////////////////////////////////////////////

import * as base from "../components/ui/base";
import * as hooks from "../components/hooks";
import { Link } from "react-router-dom";

// TODO export /////////////////////////////////////////////////////////////////////////////////////////////////////////

export default function Page(): JSX.Element {
  const [trips, setTrips, resetTrips] = hooks.useStateCustom1([]);
  const [currentTime, setCurrentTime, resetCurrentTime] = hooks.useStateCustom1(
    // @ts-ignore
    Date().toString()
  );

  async function GetTripsData() {
    const action = "monitoring";
    const token =
      "pbkdf2_sha256$390000$aWV2vUbGxC6OYns3ZQw5q5$WqX/DUYyjIjA7X42p/paeS9gYlGd3dQsGk+ZUaxdB6Y=";
    const formData = new FormData();
    const config = {
      url: `api/trips/shift/`,
      method: `GET`,
      timeout: 5000,
      headers: {
        Authorization: `action=${action};token=${token};`,
      },
      data: formData,
    };
    const response = await axios(config);
    setTrips(response.data.response);
    // @ts-ignore
    setCurrentTime(Date().toString());
  }

  async function GetMonitoring() {
    await GetTripsData();
    setTimeout(() => GetMonitoring(), 3000);
  }

  useEffect(() => {
    GetMonitoring();
  }, []);
  // TODO return ///////////////////////////////////////////////////////////////////////////////////////////////////////
  return (
    <base.Base1
      title={"Анализ рейсов"}
      description={
        "на этой странице автоматически происходит обновление показателей КПД рейсов"
      }
    >
      <div className={"lead tex-danger fw-bold display-6"}>
        {currentTime && currentTime.length > 0 && currentTime.split(" G")[0]}
      </div>
      <div className="row row-cols-sm-1 row-cols-3 row-cols-md-4 row-cols-lg-5 text-center m-0 p-0">
        <table className="table table-dark table-hover table-striped small">
          <thead>
            <tr>
              <th scope="col">№</th>
              <th scope="col">Время погрузки</th>
              <th scope="col">Время разгрузки</th>
              <th scope="col">Экскаватор</th>
              <th scope="col">Самосвал</th>
              <th scope="col">Место разгрузки</th>
              <th scope="col">Тип материала</th>
              <th scope="col">Время движения</th>
              <th scope="col">Вес</th>
              <th scope="col">Количество ковшей</th>
              <th scope="col">Средняя скорость</th>
              <th scope="col">Расстояние от погрузки</th>
              <th scope="col">Расстояние от разгрузки</th>
              <th scope="col">Высота погрузки</th>
              <th scope="col">Высота разгрузки</th>
            </tr>
          </thead>
          <tbody>
            {trips && trips.data && trips.data.length > 0
              ? trips.data.map(
                  // @ts-ignore
                  (item, index) => (
                    <tr>
                      <th># {index}</th>
                      <td>
                        {item.timeload.split("T")[0]}{" "}
                        {item.timeload.split("T")[1].split("+")[0]}
                      </td>
                      <td>
                        {item.timeunload.split("T")[0]}{" "}
                        {item.timeunload.split("T")[1].split("+")[0]}
                      </td>
                      <td>{item.shovid}</td>
                      <td>{item.vehid}</td>
                      <td>{item.unloadid}</td>
                      <td>{item.worktype}</td>
                      <td>{item.movetime.split("T")[1].split("+")[0]}</td>
                      <td>{item.weigth}</td>
                      <td>{item.bucketcount}</td>
                      <td>{item.avspeed}</td>
                      <td>{item.length}</td>
                      <td>{item.unloadlength}</td>
                      <td>{item.loadheight}</td>
                      <td>{item.unloadheight}</td>
                    </tr>
                  )
                )
              : ""}
          </tbody>
        </table>
      </div>
    </base.Base1>
  );
}
