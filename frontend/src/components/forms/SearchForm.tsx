import React, { useState } from 'react';
import { Form, InputGroup, Button, Dropdown, Badge } from 'react-bootstrap';

interface SearchFormProps {
  onSearch: (query: string, filters: Record<string, any>) => void;
  placeholder?: string;
  filters?: Array<{
    key: string;
    label: string;
    type: 'select' | 'date' | 'text';
    options?: Array<{ value: string; label: string }>;
  }>;
  className?: string;
}

const SearchForm: React.FC<SearchFormProps> = ({ 
  onSearch,
  placeholder = 'Search...',
  filters = [],
  className = ''
}) => {
  const [query, setQuery] = useState('');
  const [activeFilters, setActiveFilters] = useState<Record<string, any>>({});

  const handleSearch = (e: React.FormEvent) => {
    e.preventDefault();
    onSearch(query, activeFilters);
  };

  const handleFilterChange = (key: string, value: any) => {
    setActiveFilters(prev => ({ ...prev, [key]: value }));
  };

  const removeFilter = (key: string) => {
    const newFilters = { ...activeFilters };
    delete newFilters[key];
    setActiveFilters(newFilters);
  };

  const clearAllFilters = () => {
    setActiveFilters({});
    setQuery('');
    onSearch('', {});
  };

  return (
    <Form onSubmit={handleSearch} className={className}>
      <InputGroup className="mb-3">
        <Form.Control
          type="text"
          placeholder={placeholder}
          value={query}
          onChange={(e) => setQuery(e.target.value)}
        />
        <Button variant="outline-secondary" type="submit">
          <i className="fas fa-search me-2"></i>
          Search
        </Button>
      </InputGroup>

      <div className="d-flex flex-wrap align-items-center mb-3">
        {filters.map(filter => (
          <Dropdown key={filter.key} className="me-2 mb-2">
            <Dropdown.Toggle variant="outline-secondary" size="sm">
              <i className="fas fa-filter me-2"></i>
              {filter.label}
            </Dropdown.Toggle>
            <Dropdown.Menu>
              {filter.type === 'select' && filter.options?.map(option => (
                <Dropdown.Item 
                  key={option.value} 
                  onClick={() => handleFilterChange(filter.key, option.value)}
                >
                  {option.label}
                </Dropdown.Item>
              ))}
              {filter.type === 'date' && (
                <>
                  <Dropdown.Item onClick={() => handleFilterChange(filter.key, 'today')}>
                    Today
                  </Dropdown.Item>
                  <Dropdown.Item onClick={() => handleFilterChange(filter.key, 'this_week')}>
                    This Week
                  </Dropdown.Item>
                  <Dropdown.Item onClick={() => handleFilterChange(filter.key, 'this_month')}>
                    This Month
                  </Dropdown.Item>
                  <Dropdown.Item onClick={() => handleFilterChange(filter.key, 'this_year')}>
                    This Year
                  </Dropdown.Item>
                </>
              )}
              {filter.type === 'text' && (
                <div className="p-3">
                  <Form.Control
                    type="text"
                    placeholder={`Enter ${filter.label.toLowerCase()}...`}
                    value={activeFilters[filter.key] || ''}
                    onChange={(e) => handleFilterChange(filter.key, e.target.value)}
                  />
                </div>
              )}
            </Dropdown.Menu>
          </Dropdown>
        ))}

        {Object.keys(activeFilters).length > 0 && (
          <Button 
            variant="outline-danger" 
            size="sm" 
            className="me-2 mb-2"
            onClick={clearAllFilters}
          >
            <i className="fas fa-times-circle me-2"></i>
            Clear All
          </Button>
        )}

        <div className="d-flex flex-wrap">
          {Object.entries(activeFilters).map(([key, value]) => {
            const filter = filters.find(f => f.key === key);
            if (!filter) return null;

            let displayValue = value;
            if (filter.type === 'select' && filter.options) {
              const option = filter.options.find(o => o.value === value);
              displayValue = option ? option.label : value;
            }

            return (
              <Badge
                key={key}
                bg="primary"
                className="me-2 mb-2 d-flex align-items-center"
              >
                {filter.label}: {displayValue}
                <Button 
                  variant="link" 
                  className="p-0 ms-2 text-white"
                  onClick={() => removeFilter(key)}
                >
                  <i className="fas fa-times"></i>
                </Button>
              </Badge>
            );
          })}
        </div>
      </div>
    </Form>
  );
};

export default SearchForm;