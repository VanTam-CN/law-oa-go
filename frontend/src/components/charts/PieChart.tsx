import React from 'react';
import { Card } from 'react-bootstrap';

interface PieChartProps {
  title: string;
  data: Array<{
    label: string;
    value: number;
    color: string;
  }>;
  height?: number;
}

const PieChart: React.FC<PieChartProps> = ({ title, data, height = 300 }) => {
  const total = data.reduce((sum, item) => sum + item.value, 0);
  
  // Calculate percentages and angles for pie slices
  const chartData = data.map(item => ({
    ...item,
    percentage: total > 0 ? (item.value / total) * 100 : 0,
    angle: total > 0 ? (item.value / total) * 360 : 0
  }));
  
  return (
    <Card>
      <Card.Header>{title}</Card.Header>
      <Card.Body>
        <div className="chart-container" style={{ height: `${height}px` }}>
          {data.length > 0 && total > 0 ? (
            <div className="d-flex">
              <div className="flex-grow-1 d-flex align-items-center justify-content-center">
                <div className="pie-chart-placeholder rounded-circle bg-light d-flex align-items-center justify-content-center" style={{ width: '200px', height: '200px' }}>
                  <i className="fas fa-chart-pie fa-3x text-muted"></i>
                </div>
              </div>
              <div className="legend" style={{ width: '40%' }}>
                {chartData.map((item, index) => (
                  <div key={index} className="d-flex align-items-center mb-2">
                    <div 
                      className="rounded me-2" 
                      style={{ 
                        width: '20px', 
                        height: '20px', 
                        backgroundColor: item.color 
                      }}
                    ></div>
                    <div className="flex-grow-1">
                      <div className="small fw-bold">{item.label}</div>
                      <div className="small text-muted">
                        {item.value} ({item.percentage.toFixed(1)}%)
                      </div>
                    </div>
                  </div>
                ))}
              </div>
            </div>
          ) : (
            <div className="d-flex align-items-center justify-content-center h-100">
              <div className="text-center">
                <i className="fas fa-chart-pie fa-3x text-muted mb-3"></i>
                <p className="text-muted">No data available</p>
              </div>
            </div>
          )}
        </div>
      </Card.Body>
    </Card>
  );
};

export default PieChart;