import { Area, AreaChart, ResponsiveContainer } from 'recharts'

interface Props {
  data: Array<{ value: number }>
  color?: string
  height?: number
}

export function SparkArea({ data, color = '#2DD4BF', height = 48 }: Props) {
  if (!data || data.length === 0) return <div style={{ height }} />
  const id = `spark-${color.replace('#', '')}`
  return (
    <ResponsiveContainer width="100%" height={height}>
      <AreaChart data={data} margin={{ top: 2, right: 0, left: 0, bottom: 0 }}>
        <defs>
          <linearGradient id={id} x1="0" y1="0" x2="0" y2="1">
            <stop offset="5%" stopColor={color} stopOpacity={0.35} />
            <stop offset="95%" stopColor={color} stopOpacity={0} />
          </linearGradient>
        </defs>
        <Area
          type="monotone"
          dataKey="value"
          stroke={color}
          strokeWidth={1.5}
          fill={`url(#${id})`}
          dot={false}
          isAnimationActive={false}
        />
      </AreaChart>
    </ResponsiveContainer>
  )
}
