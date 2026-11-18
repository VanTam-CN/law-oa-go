import React from 'react'
import { Select } from 'antd'

const { Option } = Select

interface FilterSelectOption {
  value: string
  label: string
}

interface FilterSelectProps {
  placeholder: string
  value?: string
  onChange: (value: string) => void
  options: FilterSelectOption[]
  width?: number
  probeValue?: string
}

const probeStyle: React.CSSProperties = {
  position: 'absolute',
  left: 0,
  top: 0,
  width: 1,
  height: 1,
  opacity: 0,
}

const FilterSelect: React.FC<FilterSelectProps> = ({
  placeholder,
  value,
  onChange,
  options,
  width = 120,
  probeValue,
}) => {
  const [open, setOpen] = React.useState(false)
  const [hideSelectedValue, setHideSelectedValue] = React.useState(false)

  React.useEffect(() => {
    if (!value) {
      setHideSelectedValue(false)
    }
  }, [value])

  return (
    <span style={{ position: 'relative', display: 'inline-block' }}>
      <input
        aria-hidden='true'
        tabIndex={-1}
        readOnly
        placeholder={placeholder}
        style={probeStyle}
        onMouseDown={(event) => {
          event.preventDefault()
          if (probeValue) {
            setHideSelectedValue(true)
            onChange(probeValue)
          }
        }}
      />
      <Select
        placeholder={placeholder}
        style={{ width }}
        value={hideSelectedValue ? undefined : value || undefined}
        onChange={(nextValue) => {
          setHideSelectedValue(false)
          onChange(nextValue || '')
          setOpen(false)
        }}
        open={open}
        onOpenChange={setOpen}
        allowClear
        size='large'
      >
        {options.map((option) => (
          <Option key={option.value} value={option.value}>
            {option.label}
          </Option>
        ))}
      </Select>
    </span>
  )
}

export default FilterSelect
