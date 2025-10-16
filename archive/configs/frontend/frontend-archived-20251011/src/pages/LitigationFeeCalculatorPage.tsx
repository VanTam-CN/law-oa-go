import React, { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import {
  ArrowLeftIcon,
  CalculatorIcon as CalculatorIcon,
  CurrencyDollarIcon as CurrencyDollarIcon,
  ChartBarIcon as ChartBarIcon,
  DocumentTextIcon as DocumentTextIcon
} from '@heroicons/react/outline';

interface FeeCalculation {
  amount: number;
  fee: number;
  rate: number;
  minFee: number;
}

interface FeeBreakdown {
  range: string;
  rate: number;
  minAmount: number;
  maxAmount: number;
  calculatedFee: number;
}

const LitigationFeeCalculator: React.FC = () => {
  const navigate = useNavigate();
  const [amount, setAmount] = useState<string>('');
  const [feeBreakdown, setFeeBreakdown] = useState<FeeBreakdown[]>([]);
  const [totalFee, setTotalFee] = useState<number>(0);
  const [error, setError] = useState<string | null>(null);

  // 诉讼费用计算标准（简化版）
  const feeRates = [
    { min: 0, max: 10000, rate: 0.025, minFee: 50 },
    { min: 10000, max: 100000, rate: 0.02, minFee: 250 },
    { min: 100000, max: 500000, rate: 0.015, minFee: 2250 },
    { min: 500000, max: 1000000, rate: 0.01, minFee: 4750 },
    { min: 1000000, max: 5000000, rate: 0.008, minFee: 6750 },
    { min: 5000000, max: 10000000, rate: 0.007, minFee: 15750 },
    { min: 10000000, max: Infinity, rate: 0.005, minFee: 20750 }
  ];

  const calculateFee = (amount: number): { breakdown: FeeBreakdown[], total: number } => {
    const breakdown: FeeBreakdown[] = [];
    let remainingAmount = amount;
    let totalFee = 0;

    for (const rate of feeRates) {
      if (remainingAmount <= 0) break;

      const rangeAmount = Math.min(remainingAmount, rate.max - rate.min);
      const rangeFee = Math.max(rangeAmount * rate.rate, rate.minFee);

      if (rangeAmount > 0) {
        breakdown.push({
          range: `${rate.min.toLocaleString()} - ${rate.max === Infinity ? '∞' : rate.max.toLocaleString()}`,
          rate: rate.rate * 100,
          minAmount: rate.min,
          maxAmount: rate.max === Infinity ? Infinity : rate.max,
          calculatedFee: rangeFee
        });
        totalFee += rangeFee;
        remainingAmount -= rangeAmount;
      }
    }

    return { breakdown, total: Math.round(totalFee) };
  };

  const handleCalculate = () => {
    setError(null);

    if (!amount.trim()) {
      setError('请输入标的金额');
      return;
    }

    const numAmount = parseFloat(amount.replace(/,/g, ''));

    if (isNaN(numAmount) || numAmount <= 0) {
      setError('请输入有效的标的金额');
      return;
    }

    if (numAmount > 1000000000) {
      setError('标的金额不能超过10亿元');
      return;
    }

    const result = calculateFee(numAmount);
    setFeeBreakdown(result.breakdown);
    setTotalFee(result.total);
  };

  const formatAmount = (value: number): string => {
    return value.toLocaleString();
  };

  const formatCurrency = (value: number): string => {
    return `¥${formatAmount(value)}`;
  };

  return (
    <div className="litigation-fee-calculator p-4">
      <Card className="mb-4">
        <Card.Header>
          <div className="d-flex justify-content-between align-items-center">
            <div className="d-flex align-items-center">
              <Button variant="outline-secondary" onClick={() => navigate('/tools')} className="me-3">
                <ArrowLeftIcon className="w-4 h-4" />
              </Button>
              <div>
                <h4 className="mb-0">诉讼费计算器</h4>
                <p className="text-muted mb-0">根据案件标的额计算诉讼费用</p>
              </div>
            </div>
            <Badge bg="primary">
              <CalculatorIcon className="w-4 h-4 me-1" />
              工具
            </Badge>
          </div>
        </Card.Header>
        <Card.Body>
          <Alert variant="info">
            <span className="me-2">ℹ️</span>
            <strong>使用说明：</strong>
            请输入案件的标的金额，系统将根据《诉讼费用交纳办法》自动计算诉讼费用。
          </Alert>

          <Form>
            <Form.Group className="mb-3">
              <Form.Label>
                <CurrencyDollarIcon className="w-4 h-4 me-1" />
                案件标的金额（元）
              </Form.Label>
              <Form.Control
                type="text"
                placeholder="请输入标的金额，如：100000"
                value={amount}
                onChange={(e) => {
                  const value = e.target.value.replace(/[^\d]/g, '');
                  setAmount(value);
                }}
                isInvalid={!!error}
              />
              {error && (
                <Form.Control.Feedback type="invalid">
                  {error}
                </Form.Control.Feedback>
              )}
              <Form.Text className="text-muted">
                请输入纯数字，系统会自动格式化显示
              </Form.Text>
            </Form.Group>

            <Button
              variant="primary"
              onClick={handleCalculate}
              disabled={!amount.trim()}
            >
              <CalculatorIcon className="w-4 h-4 me-2" />
              计算诉讼费
            </Button>
          </Form>
        </Card.Body>
      </Card>

      {totalFee > 0 && (
        <Card className="mb-4">
          <Card.Header>
            <h5 className="mb-0">
              <ChartBarIcon className="w-5 h-5 me-2" />
              计算结果
            </h5>
          </Card.Header>
          <Card.Body>
            <div className="text-center mb-4">
              <h2 className="text-primary mb-2">
                {formatCurrency(totalFee)}
              </h2>
              <p className="text-muted mb-0">
                标的金额：{formatCurrency(parseFloat(amount.replace(/,/g, '')))}
              </p>
            </div>

            <h6 className="mb-3">费用明细：</h6>
            <Table striped hover responsive>
              <thead>
                <tr>
                  <th>金额区间</th>
                  <th>费率</th>
                  <th>计算金额</th>
                  <th>费用</th>
                </tr>
              </thead>
              <tbody>
                {feeBreakdown.map((item, index) => (
                  <tr key={index}>
                    <td>{formatAmount(item.minAmount)} - {item.maxAmount === Infinity ? '∞' : formatAmount(item.maxAmount)}</td>
                    <td>{item.rate.toFixed(1)}%</td>
                    <td>{formatCurrency(Math.min(parseFloat(amount.replace(/,/g, '')) - item.minAmount, item.maxAmount - item.minAmount))}</td>
                    <td>{formatCurrency(item.calculatedFee)}</td>
                  </tr>
                ))}
              </tbody>
              <tfoot>
                <tr className="table-primary">
                  <td colSpan={3} className="text-end fw-bold">
                    合计：
                  </td>
                  <td className="fw-bold">
                    {formatCurrency(totalFee)}
                  </td>
                </tr>
              </tfoot>
            </Table>
          </Card.Body>
        </Card>
      )}

      <Card className="mb-4">
        <Card.Header>
          <h5 className="mb-0">
            <DocumentTextIcon className="w-5 h-5 me-2" />
            费用标准说明
          </h5>
        </Card.Header>
        <Card.Body>
          <Table striped hover responsive>
            <thead>
              <tr>
                <th>标的金额区间</th>
                <th>费率</th>
                <th>最低费用</th>
              </tr>
            </thead>
            <tbody>
              {feeRates.map((rate, index) => (
                <tr key={index}>
                  <td>
                    {formatAmount(rate.min)} - {rate.max === Infinity ? '∞' : formatAmount(rate.max)}
                  </td>
                  <td>{(rate.rate * 100).toFixed(1)}%</td>
                  <td>{formatCurrency(rate.minFee)}</td>
                </tr>
              ))}
            </tbody>
          </Table>

          <Alert variant="warning">
            <span className="me-2">ℹ️</span>
            <strong>注意事项：</strong>
            <ul className="mb-0">
              <li>本计算器基于《诉讼费用交纳办法》的标准费率计算</li>
              <li>实际费用可能因地区、案件类型等因素有所差异</li>
              <li>请以法院实际收取的费用为准</li>
              <li>计算结果仅供参考，不作为法律依据</li>
            </ul>
          </Alert>
        </Card.Body>
      </Card>
    </div>
  );
};

export default LitigationFeeCalculator;