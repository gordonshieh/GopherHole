import React from 'react';
import logo from './logo.svg';
import './App.css';

import HistoryTable from './HistoryTable';
import { ThemeOptions, createMuiTheme, ThemeProvider } from '@material-ui/core';

function App() {
  const darkTheme = createMuiTheme({
    palette: {
      type: 'dark',
    },
  });
  return (
    <ThemeProvider theme={darkTheme}>
      <div className="App">
          <HistoryTable/>
      </div>
    </ThemeProvider>
  );
}

export default App;
