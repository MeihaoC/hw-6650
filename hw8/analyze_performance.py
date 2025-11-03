"""
Analyze and compare MySQL vs DynamoDB performance
Generates statistics and comparison tables for HW8 STEP III
"""

import json
import statistics

def load_results():
    try:
        with open('combined_results.json', 'r') as f:
            combined = json.load(f)
        
        # Handle both flat array and nested structure formats
        if isinstance(combined, list):
            # Flat array format (expected)
            mysql = [op for op in combined if op.get('database') == 'mysql']
            dynamodb = [op for op in combined if op.get('database') == 'dynamodb']
        else:
            # Nested structure (fallback)
            mysql = combined.get('mysql', [])
            dynamodb = combined.get('dynamodb', [])
        
        return mysql, dynamodb
    except FileNotFoundError:
        print("❌ Error: combined_results.json not found!")
        print("   Please run create_combined_results.py first")
        raise
    except json.JSONDecodeError as e:
        print(f"❌ Error parsing JSON: {e}")
        raise

def calculate_percentile(data, percentile):
    """Calculate percentile from sorted data"""
    if not data:
        return 0
    sorted_data = sorted(data)
    index = (percentile / 100) * len(sorted_data)
    if index.is_integer():
        return sorted_data[int(index) - 1]
    else:
        lower = sorted_data[int(index) - 1]
        upper = sorted_data[int(index)]
        return (lower + upper) / 2

def analyze_database(results, db_name):
    """Analyze performance for a single database"""
    # Filter successful operations only
    successful = [r for r in results if r['success']]
    
    # Overall statistics
    all_times = [r['response_time'] for r in successful]
    
    stats = {
        'database': db_name,
        'total_operations': len(results),
        'successful_operations': len(successful),
        'success_rate': (len(successful) / len(results)) * 100 if results else 0,
        'avg_response_time': statistics.mean(all_times) if all_times else 0,
        'min_response_time': min(all_times) if all_times else 0,
        'max_response_time': max(all_times) if all_times else 0,
        'median_response_time': statistics.median(all_times) if all_times else 0,
        'p50': calculate_percentile(all_times, 50),
        'p95': calculate_percentile(all_times, 95),
        'p99': calculate_percentile(all_times, 99),
    }
    
    # Per-operation statistics
    operations = {}
    for op_type in ['create_cart', 'add_items', 'get_cart']:
        op_results = [r for r in successful if r['operation'] == op_type]
        op_times = [r['response_time'] for r in op_results]
        
        operations[op_type] = {
            'count': len(op_results),
            'avg': statistics.mean(op_times) if op_times else 0,
            'min': min(op_times) if op_times else 0,
            'max': max(op_times) if op_times else 0,
            'p50': calculate_percentile(op_times, 50),
            'p95': calculate_percentile(op_times, 95),
        }
    
    stats['operations'] = operations
    return stats

def compare_databases(mysql_stats, dynamo_stats):
    """Generate comparison metrics"""
    comparison = {}
    
    # Overall comparison
    comparison['overall'] = {
        'mysql_avg': mysql_stats['avg_response_time'],
        'dynamodb_avg': dynamo_stats['avg_response_time'],
        'difference_ms': dynamo_stats['avg_response_time'] - mysql_stats['avg_response_time'],
        'difference_pct': ((dynamo_stats['avg_response_time'] - mysql_stats['avg_response_time']) / mysql_stats['avg_response_time']) * 100,
        'winner': 'MySQL' if mysql_stats['avg_response_time'] < dynamo_stats['avg_response_time'] else 'DynamoDB'
    }
    
    # Per-operation comparison
    comparison['operations'] = {}
    for op in ['create_cart', 'add_items', 'get_cart']:
        mysql_avg = mysql_stats['operations'][op]['avg']
        dynamo_avg = dynamo_stats['operations'][op]['avg']
        
        comparison['operations'][op] = {
            'mysql_avg': mysql_avg,
            'dynamodb_avg': dynamo_avg,
            'difference_ms': dynamo_avg - mysql_avg,
            'difference_pct': ((dynamo_avg - mysql_avg) / mysql_avg) * 100 if mysql_avg > 0 else 0,
            'faster': 'MySQL' if mysql_avg < dynamo_avg else 'DynamoDB'
        }
    
    return comparison

def print_comparison_table(mysql_stats, dynamo_stats, comparison):
    """Print formatted comparison tables"""
    
    print("\n" + "=" * 85)
    print("HW8 STEP III: MySQL vs DynamoDB Performance Comparison")
    print("=" * 85)
    
    # Table 1: Overall Performance
    print("\n📊 OVERALL PERFORMANCE COMPARISON")
    print("-" * 85)
    print(f"{'Metric':<30} {'MySQL':<15} {'DynamoDB':<15} {'Winner':<15} {'Margin':<15}")
    print("-" * 85)
    
    metrics = [
        ('Avg Response Time (ms)', 'avg_response_time', '{:.2f}'),
        ('P50 Response Time (ms)', 'p50', '{:.2f}'),
        ('P95 Response Time (ms)', 'p95', '{:.2f}'),
        ('P99 Response Time (ms)', 'p99', '{:.2f}'),
        ('Min Response Time (ms)', 'min_response_time', '{:.2f}'),
        ('Max Response Time (ms)', 'max_response_time', '{:.2f}'),
        ('Success Rate (%)', 'success_rate', '{:.1f}'),
    ]
    
    for label, key, fmt in metrics:
        mysql_val = mysql_stats[key]
        dynamo_val = dynamo_stats[key]
        
        if key == 'success_rate':
            winner = 'Tie' if mysql_val == dynamo_val else ('MySQL' if mysql_val > dynamo_val else 'DynamoDB')
            margin = f"{abs(mysql_val - dynamo_val):.1f}%"
        else:
            winner = 'MySQL' if mysql_val < dynamo_val else 'DynamoDB'
            margin = f"{abs(dynamo_val - mysql_val):.2f}ms"
        
        print(f"{label:<30} {fmt.format(mysql_val):<15} {fmt.format(dynamo_val):<15} {winner:<15} {margin:<15}")
    
    print("-" * 85)
    print(f"{'Total Operations':<30} {mysql_stats['total_operations']:<15} {dynamo_stats['total_operations']:<15}")
    print("-" * 85)
    
    # Table 2: Operation-Specific Breakdown
    print("\n📋 OPERATION-SPECIFIC COMPARISON")
    print("-" * 85)
    print(f"{'Operation':<20} {'MySQL Avg':<15} {'DynamoDB Avg':<15} {'Faster':<15} {'Margin':<15}")
    print("-" * 85)
    
    for op in ['create_cart', 'add_items', 'get_cart']:
        op_comp = comparison['operations'][op]
        op_label = op.replace('_', ' ').title()
        
        print(f"{op_label:<20} {op_comp['mysql_avg']:.2f}ms{'':<8} "
              f"{op_comp['dynamodb_avg']:.2f}ms{'':<8} "
              f"{op_comp['faster']:<15} "
              f"{abs(op_comp['difference_ms']):.2f}ms ({abs(op_comp['difference_pct']):.1f}%)")
    
    print("-" * 85)
    
    # Summary
    overall = comparison['overall']
    print("\n🎯 SUMMARY")
    print("-" * 85)
    print(f"Overall Winner: {overall['winner']}")
    print(f"Average Performance Difference: {abs(overall['difference_ms']):.2f}ms ({abs(overall['difference_pct']):.1f}%)")
    print(f"Verdict: Performance is {'statistically equivalent' if abs(overall['difference_pct']) < 5 else 'significantly different'}")
    print("-" * 85)

def save_analysis(mysql_stats, dynamo_stats, comparison):
    """Save analysis results to JSON"""
    analysis = {
        'mysql': mysql_stats,
        'dynamodb': dynamo_stats,
        'comparison': comparison
    }
    
    with open('performance_analysis.json', 'w') as f:
        json.dump(analysis, f, indent=2)
    
    print("\n💾 Detailed analysis saved to: performance_analysis.json")

def main():
    # Load results
    print("📂 Loading test results...")
    mysql_results, dynamodb_results = load_results()
    
    # Analyze each database
    print("📊 Analyzing MySQL performance...")
    mysql_stats = analyze_database(mysql_results, 'MySQL')
    
    print("📊 Analyzing DynamoDB performance...")
    dynamo_stats = analyze_database(dynamodb_results, 'DynamoDB')
    
    # Compare
    print("⚖️  Comparing performance...")
    comparison = compare_databases(mysql_stats, dynamo_stats)
    
    # Print results
    print_comparison_table(mysql_stats, dynamo_stats, comparison)
    
    # Save analysis
    save_analysis(mysql_stats, dynamo_stats, comparison)
    
    print("\n✅ Analysis complete!")

if __name__ == "__main__":
    main()