
import React, { Component } from 'react';
import {Table, TableBody, TableCell, TableContainer, TableHead, TableRow, Paper }  from '@material-ui/core';
import { isCompositeComponent } from 'react-dom/test-utils';

type History = {
    type: string,
    source: string,
    host: string,
    timestamp: Date,
    block: boolean
}

type HistoryState = {
    isLoaded: boolean,
    items: Array<History>,
    date: string,
    timerId?: NodeJS.Timeout
}

function formatDate(date: Date): String {
    let formatted = String(date.getFullYear());
    let month = date.getMonth();
    let day = date.getDate();
    let hours = date.getHours();
    let minutes = date.getMinutes();
    let seconds = date.getSeconds();

    formatted += "-";
    if (month < 9) {
        formatted += "0";
    }
    formatted += String(month);

    formatted += "-";
    if (day < 9) {
        formatted += "0";
    }
    formatted += String(day);

    formatted += " ";
    if (hours < 9) {
        formatted += "0";
    }
    formatted += String(hours);

    formatted += ":";
    if (minutes < 9) {
        formatted += "0";
    }
    formatted += String(minutes);

    formatted += ":";
    if (seconds < 9) {
        formatted += "0";
    }
    formatted += String(seconds);
    return formatted;
}

class HistoryTable extends Component<{}, HistoryState> {
    ws = new WebSocket('ws://localhost:1323/history-stream');

    constructor() {
        super({});
        this.state = {isLoaded: false, items: [], date: new Date().toLocaleTimeString()};
    }

    private handleClick = (event: React.MouseEvent<HTMLTableElement>) => {
    
        this.setState(({ date }) => ({
          date: "abc"
        }));
      };
    

    tick() {
        this.setState({date: new Date().toLocaleTimeString()})
    }
    
    componentDidMount() {
        this.setState({timerId: setInterval(
                () => this.tick(),
                1000
            )}
        );
        fetch("http://localhost:1323/history")
            .then(res => res.json())
            .then(
                (result) => {
                    this.setState({items: result})
                },
                (error) => {
                    console.log(error);
                }
            );
        this.ws.onmessage = (ev: MessageEvent) => {
            let currResults = this.state.items;
            currResults.unshift(JSON.parse(ev.data));
            this.setState({items: currResults.slice(0, 50)})
        };
    }

    render() {
        return (
            <TableContainer component={Paper}>
                <Table>
                    <TableHead>
                        <TableRow>
                            <TableCell>Time</TableCell>
                            <TableCell>Type</TableCell>
                            <TableCell>Domain</TableCell>
                            <TableCell>Client</TableCell>
                            <TableCell>Blocked</TableCell>
                        </TableRow>
                    </TableHead>
                    <TableBody>
                        {this.state.items.map(row => {
                            let dateObj = new Date(row.timestamp);
                            return (
                            <TableRow>
                                <TableCell>{formatDate(dateObj)}</TableCell>
                                <TableCell>{row.type}</TableCell>
                                <TableCell>{row.host}</TableCell>
                                <TableCell>{row.source}</TableCell>
                                <TableCell>{row.block ? "Blocked" : "Ok"}</TableCell>
                            </TableRow>)
                        })}
                    </TableBody>
                </Table>
            </TableContainer>
        )
    }
}

export default HistoryTable;
