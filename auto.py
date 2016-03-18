from time import sleep
import subprocess

def main():
  files = subprocess.check_output("git ls-files . --others --exclude-standard", shell=True).strip().split("\n")

  total = len(files)
  for i in range(total):
    j = i + 1
    if j % 5 == 0:
      cmd = "git add %s %s %s %s %s; git commit -m 'left %s'; git push;" % (files[i-4], files[i-3], files[i-2], files[i-1], files[i], total - i)
      subprocess.call(cmd, shell=True)
      sleep(.5)

if __name__ == "__main__":
	main()
