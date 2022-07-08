import "./App.css";

import React from "react";
import Tiptap from "./Tiptap.jsx";
import Navigator from "./Navigator";
// import './Tiptap.css';

const App = () => {
  return (
    <div className="App">
      <Navigator/>
      <Tiptap />
    </div>
  );
};

export default App;