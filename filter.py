import pandas as pd

# Read the CSV file
df = pd.read_csv('airport_data.csv')

# Filter the DataFrame to include only large airports
large_airports_df = df[df['type'] == 'large_airport']

# Write the filtered DataFrame to a new CSV file
large_airports_df.to_csv('large_airports.csv', index=False)