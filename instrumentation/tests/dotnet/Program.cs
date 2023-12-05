using System;
using System.Collections;

foreach (DictionaryEntry de in Environment.GetEnvironmentVariables())
    Console.WriteLine("{0}={1}", de.Key, de.Value);