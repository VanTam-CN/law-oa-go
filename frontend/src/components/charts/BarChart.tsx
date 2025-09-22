import React from 'react';
import { Card } from 'react-bootstrap';

interface BarChartProps {
  title: string;
  data: Array<{
    label: string;
    value: number;
    color: string;
  }>;
  height?: number;
}

const BarChart: React.FC<BarChartProps> = ({ title, data, height = 300 }) => {
  const maxValue = Math.max(...data.map(item => item.value), 1);
  
  return (
    <Card>
      <Card.Header>{title}</Card.Header>
      <Card.Body>
        <div className="chart-container" style={{ height: `${height}px` }}>
          {data.length > 0 ? (
            <div className="d-flex align-items-end h-100">
              {data.map((item, index) => (
                <div 
                  key={index} 
                  className="d-flex flex-column align-items-center mx-1 flex-grow-1"
                  style={{ height: '100%' }}
                >
                  <div 
                    className="w-100 rounded-top"
                    style={{ 
                      height: `${(item.value / maxValue) * 100}%`,
                      backgroundColor: item.color
                    }}
                  ></div>
                  <div className="mt-2 small text-center" style={{ width: '100%' }}>
                    <div className="fw-bold">{item.value}</div>
                    <div className="text-muted">{item.label}</div>
                  </div>
                </div>
              ))}
            </div>
          ) : (
            <div className="d-flex align-items-center justify-content-center h-100">
              <div className="text-center">
                <i className="fas fa-chart-bar fa-3x text-muted mb-3"></i>
                <p className="text-muted">No data available</p>
              </div>
            </div>
          )}
        </div>
      </Card.Body>
    </Card>
  );
};

export default BarChart;