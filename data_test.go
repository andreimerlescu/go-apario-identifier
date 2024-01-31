package go_apario_identifier

import (
	`testing`
)

func Benchmark_fibonacci(b *testing.B) {
	b.Run("rFibonacci", func(b *testing.B) {
		for j := 0; j < 33; j++ {
			rFibonacci(j)
		}
	})
	b.Run("fibonacci", func(b *testing.B) {
		for j := 0; j < 33; j++ {
			fibonacci(j)
		}
	})

}

func Test_rFibonacci(t *testing.T) {
	type args struct {
		n int
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{
			name: "fib(1)",
			args: args{n: 1},
			want: 1,
		},
		{
			name: "fib(2)",
			args: args{n: 2},
			want: 1,
		},
		{
			name: "fib(3)",
			args: args{n: 3},
			want: 2,
		},
		{
			name: "fib(4)",
			args: args{n: 4},
			want: 3,
		},
		{
			name: "fib(5)",
			args: args{n: 5},
			want: 5,
		},
		{
			name: "fib(6)",
			args: args{n: 6},
			want: 8,
		},
		{
			name: "fib(7)",
			args: args{n: 7},
			want: 13,
		},
		{
			name: "fib(8)",
			args: args{n: 8},
			want: 21,
		},
		{
			name: "fib(9)",
			args: args{n: 9},
			want: 34,
		},
		{
			name: "fib(10)",
			args: args{n: 10},
			want: 55,
		},
		{
			name: "fib(11)",
			args: args{n: 11},
			want: 89,
		},
		{
			name: "fib(12)",
			args: args{n: 12},
			want: 144,
		},
		{
			name: "fib(13)",
			args: args{n: 13},
			want: 233,
		},
		{
			name: "fib(14)",
			args: args{n: 14},
			want: 377,
		},
		{
			name: "fib(15)",
			args: args{n: 15},
			want: 610,
		},
		{
			name: "fib(16)",
			args: args{n: 16},
			want: 987,
		},
		{
			name: "fib(17)",
			args: args{n: 17},
			want: 1597,
		},
		{
			name: "fib(18)",
			args: args{n: 18},
			want: 2584,
		},
		{
			name: "fib(19)",
			args: args{n: 19},
			want: 4181,
		},
		{
			name: "fib(20)",
			args: args{n: 20},
			want: 6765,
		},
		{
			name: "fib(21)",
			args: args{n: 21},
			want: 10946,
		},
		{
			name: "fib(22)",
			args: args{n: 22},
			want: 17711,
		},
		{
			name: "fib(23)",
			args: args{n: 23},
			want: 28657,
		},
		{
			name: "fib(24)",
			args: args{n: 24},
			want: 46368,
		},
		{
			name: "fib(25)",
			args: args{n: 25},
			want: 75025,
		},
		{
			name: "fib(26)",
			args: args{n: 26},
			want: 121393,
		},
		{
			name: "fib(27)",
			args: args{n: 27},
			want: 196418,
		},
		{
			name: "fib(28)",
			args: args{n: 28},
			want: 317811,
		},
		{
			name: "fib(29)",
			args: args{n: 29},
			want: 514229,
		},
		{
			name: "fib(30)",
			args: args{n: 30},
			want: 832040,
		},
		{
			name: "fib(31)",
			args: args{n: 31},
			want: 1346269,
		},
		{
			name: "fib(32)",
			args: args{n: 32},
			want: 2178309,
		},
		{
			name: "fib(33)",
			args: args{n: 33},
			want: 3524578,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := rFibonacci(tt.args.n); got != tt.want {
				t.Errorf("rFibonacci() = %v, want %v", got, tt.want)
			}
		})
	}
}
